package authmodule

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang/config"
	"golang/database"
	"golang/pkg/auth"
	"golang/pkg/errors"
	"golang/pkg/logging"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"golang.org/x/crypto/bcrypt"
)

// AuthService contains the authentication business logic.
type AuthService struct {
	db         *database.PostgresDB
	jwtService *auth.JWTService
	cfg        *config.Config
	httpClient *http.Client
}

// NewAuthService builds the auth service.
func NewAuthService(db *database.PostgresDB, jwtService *auth.JWTService, cfg ...*config.Config) *AuthService {
	var resolvedCfg *config.Config
	if len(cfg) > 0 {
		resolvedCfg = cfg[0]
	}
	return &AuthService{
		db:         db,
		jwtService: jwtService,
		cfg:        resolvedCfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type authUser struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Email        string     `db:"email"`
	Password     *string    `db:"password"`
	AvatarURL    *string    `db:"avatar_url"`
	AuthProvider *string    `db:"auth_provider"`
	GoogleID     *string    `db:"google_id"`
	AppleID      *string    `db:"apple_id"`
	IsActive     bool       `db:"is_active"`
	CreatedAt    *time.Time `db:"created_at"`
}

func (u *authUser) toPayload() UserPayload {
	provider := "email"
	if u.AuthProvider != nil && *u.AuthProvider != "" {
		provider = *u.AuthProvider
	}
	avatar := ""
	if u.AvatarURL != nil {
		avatar = *u.AvatarURL
	}
	return UserPayload{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		AvatarURL:    avatar,
		AuthProvider: provider,
		CreatedAt:    derefTime(u.CreatedAt),
	}
}

// Register creates a user and returns the auth response.
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.InternalError("Failed to hash password", err)
	}

	user, err := database.TransactCtx(ctx, s.db, func(tx *sqlx.Tx) (*authUser, error) {
		uid := uuid.New().String()
		var u authUser
		err := tx.QueryRowxContext(
			ctx,
			`INSERT INTO users (id, name, email, password, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW())
			 RETURNING id, name, email, created_at`,
			uid, req.Name, req.Email, string(hashedPassword),
		).StructScan(&u)
		if err != nil {
			return nil, errors.ParseDatabaseError(err)
		}
		return &u, nil
	})
	if err != nil {
		return nil, err
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User:   user.toPayload(),
		Tokens: newTokenPair(tokens),
	}, nil
}

// Login validates credentials and issues a token pair.
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	var u authUser
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		return tx.GetContext(ctx, &u,
			"SELECT id, name, email, password, avatar_url, auth_provider, is_active, created_at FROM users WHERE email = $1",
			req.Email)
	})
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.Unauthorized("Invalid email or password")
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}
	if !u.IsActive {
		return nil, errors.Forbidden("Account is deactivated")
	}
	if u.Password == nil || *u.Password == "" {
		return nil, errors.Unauthorized("Please continue with Google or Apple Sign-In for this account")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(req.Password)); err != nil {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	_, _ = s.db.DB.ExecContext(ctx, "UPDATE users SET last_login = NOW(), updated_at = NOW() WHERE id = $1", u.ID)

	tokens, err := s.jwtService.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User:   u.toPayload(),
		Tokens: newTokenPair(tokens),
	}, nil
}

// RefreshToken exchanges a refresh token for a new pair.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	tokens, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return newTokenPair(tokens), nil
}

// GetCurrentUser fetches the authenticated user profile.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID string) (*UserPayload, error) {
	var u authUser
	err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		return tx.GetContext(ctx, &u, "SELECT id, name, email, avatar_url, auth_provider, created_at FROM users WHERE id = $1", userID)
	})
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NotFound("User not found")
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}

	payload := u.toPayload()
	return &payload, nil
}

// UpdateProfile validates and updates the current user's profile.
type UpdateProfileRequest struct {
	Name string `json:"name" validate:"omitempty,min=3,max=100"`
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*UserPayload, error) {
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2", req.Name, userID)
		return err
	})
	if err != nil {
		return nil, errors.DatabaseError("Failed to update profile", err)
	}
	return s.GetCurrentUser(ctx, userID)
}

// ChangePasswordRequest payload.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
}

func (s *AuthService) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	err := s.db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var hash *string
		err := tx.GetContext(ctx, &hash, "SELECT password FROM users WHERE id = $1", userID)
		if err != nil {
			return errors.DatabaseError("Failed to fetch user", err)
		}
		if hash == nil || *hash == "" {
			return errors.BadRequest("Social sign-in accounts do not have a password set")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(*hash), []byte(req.CurrentPassword)); err != nil {
			return errors.Unauthorized("Current password is incorrect")
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return errors.InternalError("Failed to hash password", err)
		}

		_, err = tx.ExecContext(ctx, "UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2",
			string(hashed), userID)
		return err
	})
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return appErr
		}
		return errors.DatabaseError("Failed to update password", err)
	}

	// Force re-login everywhere.
	if err := s.jwtService.RevokeAllUserTokens(userID); err != nil {
		logging.Warn("failed to revoke tokens after password change",
			slog.String("user_id", userID), logging.Err(err))
	}
	return nil
}

// newTokenPair converts the underlying token pair.
func newTokenPair(t *auth.TokenPair) *TokenPair {
	if t == nil {
		return nil
	}
	return &TokenPair{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    t.ExpiresAt,
		TokenType:    t.TokenType,
	}
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Aud           string `json:"aud"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *AuthService) verifyGoogleAuth(ctx context.Context, req *OAuthLoginRequest) (*googleUserInfo, error) {
	if req.Code != "" {
		if s.cfg != nil && s.cfg.OAuth.GoogleClientID != "" && s.cfg.OAuth.GoogleClientSecret != "" {
			form := url.Values{}
			form.Set("code", req.Code)
			form.Set("client_id", s.cfg.OAuth.GoogleClientID)
			form.Set("client_secret", s.cfg.OAuth.GoogleClientSecret)
			redirectURI := s.cfg.OAuth.GoogleRedirectURI
			if redirectURI == "" {
				redirectURI = "postmessage"
			}
			form.Set("redirect_uri", redirectURI)
			form.Set("grant_type", "authorization_code")

			tokenReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
			tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := s.httpClient.Do(tokenReq)
			if err != nil {
				return nil, errors.BadRequest("Failed to communicate with Google token endpoint: " + err.Error())
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				return nil, errors.BadRequest("Google authorization code exchange failed")
			}

			var tokenRes struct {
				AccessToken string `json:"access_token"`
			}
			if err := json.Unmarshal(body, &tokenRes); err != nil {
				return nil, errors.BadRequest("Failed to parse Google token response")
			}

			reqUser, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
			reqUser.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
			userResp, err := s.httpClient.Do(reqUser)
			if err != nil {
				return nil, errors.BadRequest("Failed to fetch Google user profile: " + err.Error())
			}
			defer userResp.Body.Close()
			userBody, _ := io.ReadAll(userResp.Body)

			var info googleUserInfo
			if err := json.Unmarshal(userBody, &info); err != nil {
				return nil, errors.BadRequest("Failed to decode Google user profile")
			}
			if info.Email == "" {
				return nil, errors.BadRequest("Google account did not return an email address")
			}
			return &info, nil
		}
	}

	if req.IDToken != "" {
		tokeninfoURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(req.IDToken)
		httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, tokeninfoURL, nil)
		resp, err := s.httpClient.Do(httpReq)
		if err != nil {
			return nil, errors.Unauthorized("Failed to verify Google ID token: " + err.Error())
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, errors.Unauthorized("Google ID token validation failed")
		}

		var info googleUserInfo
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, errors.Unauthorized("Failed to decode Google tokeninfo")
		}
		if info.Email == "" {
			return nil, errors.BadRequest("Google ID token does not contain an email address")
		}
		return &info, nil
	}

	if s.cfg != nil && !s.cfg.IsProduction() && req.Email != "" {
		return &googleUserInfo{
			Sub:     "dev-" + req.Email,
			Email:   req.Email,
			Name:    req.Name,
			Picture: req.AvatarURL,
		}, nil
	}

	return nil, errors.BadRequest("Missing Google ID token or authorization code")
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

var (
	appleKeysCache     *appleJWKS
	appleKeysCacheTime time.Time
	appleKeysMu        sync.RWMutex
)

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		nBytes, err = base64.URLEncoding.DecodeString(nStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode modulus: %w", err)
		}
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		eBytes, err = base64.URLEncoding.DecodeString(eStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode exponent: %w", err)
		}
	}
	eInt := int(new(big.Int).SetBytes(eBytes).Int64())
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

func (s *AuthService) fetchAppleJWKS(ctx context.Context) (*appleJWKS, error) {
	appleKeysMu.RLock()
	if appleKeysCache != nil && time.Since(appleKeysCacheTime) < 1*time.Hour {
		cached := appleKeysCache
		appleKeysMu.RUnlock()
		return cached, nil
	}
	appleKeysMu.RUnlock()

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple public keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Apple JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks appleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode Apple JWKS: %w", err)
	}

	appleKeysMu.Lock()
	appleKeysCache = &jwks
	appleKeysCacheTime = time.Now()
	appleKeysMu.Unlock()

	return &jwks, nil
}

type appleUserInfo struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

func (s *AuthService) getAllowedAppleAudiences() []string {
	if s.cfg == nil {
		return []string{"cloud.eka-dev.eka-hub"}
	}
	var audiences []string
	for _, raw := range []string{s.cfg.OAuth.AppleClientID, s.cfg.OAuth.AppleBundleID} {
		for _, id := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				audiences = append(audiences, trimmed)
			}
		}
	}
	if len(audiences) == 0 {
		audiences = append(audiences, "cloud.eka-dev.eka-hub")
	}
	return audiences
}

func (s *AuthService) verifyAppleAuth(ctx context.Context, req *OAuthLoginRequest) (*appleUserInfo, error) {
	var email, name, sub string

	if req.IDToken != "" {
		jwks, err := s.fetchAppleJWKS(ctx)
		if err == nil && jwks != nil {
			token, parseErr := jwt.Parse(req.IDToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				kid, _ := token.Header["kid"].(string)
				for _, key := range jwks.Keys {
					if key.Kid == kid {
						return parseRSAPublicKey(key.N, key.E)
					}
				}
				return nil, fmt.Errorf("key with kid '%s' not found in Apple JWKS", kid)
			})

			if parseErr == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if iss, ok := claims["iss"].(string); !ok || iss != "https://appleid.apple.com" {
						return nil, errors.Unauthorized("Invalid Apple token issuer")
					}
					allowed := s.getAllowedAppleAudiences()
					if aud, ok := claims["aud"].(string); !ok || !containsString(allowed, aud) {
						return nil, errors.Unauthorized("Apple token audience mismatch")
					}
					if subClaim, ok := claims["sub"].(string); ok {
						sub = subClaim
					}
					if emailClaim, ok := claims["email"].(string); ok && emailClaim != "" {
						email = emailClaim
					}
				}
			} else if s.cfg != nil && s.cfg.IsProduction() {
				return nil, errors.Unauthorized("Invalid or expired Apple ID token")
			}
		}

		if (sub == "" || email == "") && (s.cfg == nil || !s.cfg.IsProduction()) {
			parser := jwt.NewParser()
			if token, _, err := parser.ParseUnverified(req.IDToken, jwt.MapClaims{}); err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub == "" {
						sub, _ = claims["sub"].(string)
					}
					if email == "" {
						email, _ = claims["email"].(string)
					}
					if name == "" {
						name, _ = claims["name"].(string)
					}
				}
			}
		}
	}

	if email == "" && req.Email != "" {
		email = req.Email
	}
	if name == "" && req.Name != "" {
		name = req.Name
	}
	if sub == "" {
		if req.Code != "" {
			sub = req.Code
		} else if email != "" {
			sub = "apple-" + email
		}
	}

	// Note: Subsequent Apple Sign-Ins omit email in ASAuthorizationAppleIDCredential
	// when user previously signed in, so look up by sub is handled in OAuthLogin!
	if sub == "" && email == "" {
		return nil, errors.BadRequest("Apple sign-in requires valid identity token or email")
	}

	return &appleUserInfo{
		Sub:     sub,
		Email:   email,
		Name:    name,
		Picture: req.AvatarURL,
	}, nil
}

func containsString(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// OAuthLogin authenticates or registers a user via Google or Apple Sign-In.
func (s *AuthService) OAuthLogin(ctx context.Context, req *OAuthLoginRequest) (*AuthResponse, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		provider = "google"
	}

	var email, name, avatarURL, providerID string
	switch provider {
	case "google":
		gInfo, err := s.verifyGoogleAuth(ctx, req)
		if err != nil {
			return nil, err
		}
		email = gInfo.Email
		name = gInfo.Name
		avatarURL = gInfo.Picture
		providerID = gInfo.Sub
	case "apple":
		aInfo, err := s.verifyAppleAuth(ctx, req)
		if err != nil {
			return nil, err
		}
		email = aInfo.Email
		name = aInfo.Name
		avatarURL = aInfo.Picture
		providerID = aInfo.Sub
	default:
		return nil, errors.BadRequest("Unsupported OAuth provider: " + provider)
	}

	providerCol := "google_id"
	if provider == "apple" {
		providerCol = "apple_id"
	}

	var u authUser
	found := false

	if providerID != "" {
		q := fmt.Sprintf("SELECT id, name, email, password, avatar_url, auth_provider, google_id, apple_id, is_active, created_at FROM users WHERE %s = $1", providerCol)
		err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			return tx.GetContext(ctx, &u, q, providerID)
		})
		if err == nil {
			found = true
		} else if !stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.DatabaseError("Database error during OAuth provider lookup", err)
		}
	}

	if !found && email != "" {
		err := s.db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
			return tx.GetContext(ctx, &u, "SELECT id, name, email, password, avatar_url, auth_provider, google_id, apple_id, is_active, created_at FROM users WHERE email = $1", email)
		})
		if err == nil {
			found = true
		} else if !stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.DatabaseError("Database error during OAuth email lookup", err)
		}
	}

	if !found {
		if email == "" {
			return nil, errors.BadRequest("Email is required for new OAuth user registration")
		}
		if name == "" {
			name = strings.Split(email, "@")[0]
		}

		uid := uuid.New().String()
		var googleID, appleID *string
		if provider == "apple" {
			appleID = &providerID
		} else {
			googleID = &providerID
		}

		createdUser, err := database.TransactCtx(ctx, s.db, func(tx *sqlx.Tx) (*authUser, error) {
			var created authUser
			err := tx.QueryRowxContext(
				ctx,
				`INSERT INTO users (id, name, email, avatar_url, auth_provider, provider_id, google_id, apple_id, is_active, last_login, created_at, updated_at)
				 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8, TRUE, NOW(), NOW(), NOW())
				 RETURNING id, name, email, avatar_url, auth_provider, google_id, apple_id, is_active, created_at`,
				uid, name, email, avatarURL, provider, providerID, googleID, appleID,
			).StructScan(&created)
			if err != nil {
				return nil, errors.ParseDatabaseError(err)
			}
			return &created, nil
		})
		if err != nil {
			return nil, err
		}
		u = *createdUser
	} else {
		if !u.IsActive {
			return nil, errors.Forbidden("Account is deactivated")
		}
		updateQ := fmt.Sprintf(
			"UPDATE users SET avatar_url = COALESCE(NULLIF($1, ''), avatar_url), auth_provider = $2, provider_id = COALESCE(NULLIF($3, ''), provider_id), %s = COALESCE(NULLIF($3, ''), %s), last_login = NOW(), updated_at = NOW() WHERE id = $4",
			providerCol, providerCol,
		)
		_, _ = s.db.DB.ExecContext(ctx, updateQ, avatarURL, provider, providerID, u.ID)
		if avatarURL != "" && (u.AvatarURL == nil || *u.AvatarURL == "") {
			u.AvatarURL = &avatarURL
		}
		u.AuthProvider = &provider
	}

	tokens, err := s.jwtService.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User:   u.toPayload(),
		Tokens: newTokenPair(tokens),
	}, nil
}
