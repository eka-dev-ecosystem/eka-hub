package authmodule

import "time"

// RegisterRequest payload.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// LoginRequest payload.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// OAuthLoginRequest handles Google and Apple Sign-In authentication.
type OAuthLoginRequest struct {
	Provider  string `json:"provider" validate:"required,oneof=google apple"`
	IDToken   string `json:"id_token"`
	Code      string `json:"code"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// RefreshTokenRequest payload.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UserPayload is a user summary for auth responses.
type UserPayload struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	AuthProvider string    `json:"auth_provider,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// TokenPair mirrors the underlying token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// AuthResponse is returned by register/login.
type AuthResponse struct {
	User   UserPayload `json:"user"`
	Tokens *TokenPair  `json:"tokens"`
}
