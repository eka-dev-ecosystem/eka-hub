package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang/cache"
	customErrors "golang/pkg/errors"
	"golang/pkg/logging"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType represents the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessSecret         string
	RefreshSecret        string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	TokenID   string    `json:"token_id"` // JTI for token revocation
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
// redis is optional: when nil, token revocation tracking is skipped and
// every token is treated as valid until it expires.
type JWTService struct {
	config *JWTConfig
	redis  *cache.RedisCache
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// NewJWTService creates a new JWT service. redis may be nil.
func NewJWTService(config *JWTConfig, redis *cache.RedisCache) *JWTService {
	return &JWTService{
		config: config,
		redis:  redis,
	}
}

// hasRedis reports whether token revocation storage is available.
func (j *JWTService) hasRedis() bool {
	return j.redis != nil
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(userID, email string) (*TokenPair, error) {
	// Generate access token
	accessToken, accessExp, err := j.generateToken(userID, email, AccessToken)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, _, err := j.generateToken(userID, email, RefreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    "Bearer",
	}, nil
}

// generateToken generates a JWT token
func (j *JWTService) generateToken(userID, email string, tokenType TokenType) (string, time.Time, error) {
	var (
		secret     string
		duration   time.Duration
		expiration time.Time
	)

	// Set secret and duration based on token type
	if tokenType == AccessToken {
		secret = j.config.AccessSecret
		duration = j.config.AccessTokenDuration
	} else {
		secret = j.config.RefreshSecret
		duration = j.config.RefreshTokenDuration
	}

	expiration = time.Now().Add(duration)
	tokenID := uuid.New().String()

	// Create claims
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.config.Issuer,
			Subject:   userID,
			ID:        tokenID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	// Store token ID in Redis for revocation tracking
	if j.hasRedis() {
		ctx := context.Background()
		redisKey := fmt.Sprintf("token:%s:%s", tokenType, tokenID)
		if err := j.redis.Set(ctx, redisKey, userID, duration); err != nil {
			// Log error but don't fail token generation
			logging.Warn("failed to store token in redis",
				slog.String("token_type", string(tokenType)),
				logging.Err(err),
			)
		}
	}

	return tokenString, expiration, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string, tokenType TokenType) (*JWTClaims, error) {
	// Choose secret based on token type
	secret := j.config.AccessSecret
	if tokenType == RefreshToken {
		secret = j.config.RefreshSecret
	}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, customErrors.Unauthorized("Invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, customErrors.Unauthorized("Invalid token claims")
	}

	// Verify token type
	if claims.TokenType != tokenType {
		return nil, customErrors.Unauthorized("Invalid token type")
	}

	// Check if token is revoked
	if j.hasRedis() {
		ctx := context.Background()
		redisKey := fmt.Sprintf("token:%s:%s", tokenType, claims.TokenID)
		exists, err := j.redis.Exists(ctx, redisKey)
		if err != nil || exists == 0 {
			return nil, customErrors.Unauthorized("Token has been revoked")
		}
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (j *JWTService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := j.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email)
}

// RevokeToken revokes a token by removing it from Redis
func (j *JWTService) RevokeToken(tokenString string, tokenType TokenType) error {
	// Parse token without validation (to get TokenID even if expired)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return customErrors.BadRequest("Invalid token format")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return customErrors.BadRequest("Invalid token claims")
	}

	if !j.hasRedis() {
		return customErrors.InternalError("Token revocation unavailable", nil)
	}

	// Remove token from Redis
	ctx := context.Background()
	redisKey := fmt.Sprintf("token:%s:%s", tokenType, claims.TokenID)
	err = j.redis.Delete(ctx, redisKey)
	if err != nil {
		return customErrors.InternalError("Failed to revoke token", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (j *JWTService) RevokeAllUserTokens(userID string) error {
	if !j.hasRedis() {
		return customErrors.InternalError("Token revocation unavailable", nil)
	}

	// This is a simplified version
	// In production, you might want to track all user tokens in a Redis set
	ctx := context.Background()

	// Store user in revoked list with expiration matching longest token duration
	revokeKey := fmt.Sprintf("user:revoked:%s", userID)
	duration := j.config.RefreshTokenDuration
	if j.config.AccessTokenDuration > duration {
		duration = j.config.AccessTokenDuration
	}

	err := j.redis.Set(ctx, revokeKey, time.Now().Unix(), duration)
	if err != nil {
		return customErrors.InternalError("Failed to revoke user tokens", err)
	}

	return nil
}

// IsUserRevoked checks if all user tokens are revoked
func (j *JWTService) IsUserRevoked(userID string) bool {
	if !j.hasRedis() {
		return false
	}

	ctx := context.Background()
	revokeKey := fmt.Sprintf("user:revoked:%s", userID)
	exists, _ := j.redis.Exists(ctx, revokeKey)
	return exists > 0
}

// GetClaimsFromToken extracts claims from token without full validation
// Useful for logging or debugging
func (j *JWTService) GetClaimsFromToken(tokenString string) (*JWTClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}
