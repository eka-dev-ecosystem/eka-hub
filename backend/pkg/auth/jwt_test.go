package auth

import (
	"testing"
	"time"

	customErrors "golang/pkg/errors"
)

func newTestJWTService() *JWTService {
	return NewJWTService(&JWTConfig{
		AccessSecret:         "test-access-secret-very-long",
		RefreshSecret:        "test-refresh-secret-very-long",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}, nil)
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	svc := newTestJWTService()

	pair, err := svc.GenerateTokenPair("user-123", "user@example.com")
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	claims, err := svc.ValidateToken(pair.AccessToken, AccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("claims.UserID = %q, want user-123", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("claims.Email = %q", claims.Email)
	}
	if claims.TokenType != AccessToken {
		t.Errorf("claims.TokenType = %q, want access", claims.TokenType)
	}
}

func TestValidateTokenRejectsWrongType(t *testing.T) {
	svc := newTestJWTService()
	pair, _ := svc.GenerateTokenPair("u1", "e@x.com")

	_, err := svc.ValidateToken(pair.RefreshToken, AccessToken)
	if err == nil {
		t.Fatal("expected error when validating refresh token as access")
	}
	if appErr, ok := err.(*customErrors.AppError); !ok || appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED AppError, got %#v", err)
	}
}

func TestValidateTokenRejectsTampered(t *testing.T) {
	svc := newTestJWTService()
	pair, _ := svc.GenerateTokenPair("u1", "e@x.com")

	tampered := pair.AccessToken[:len(pair.AccessToken)-2] + "zz"
	if _, err := svc.ValidateToken(tampered, AccessToken); err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	svc := newTestJWTService()
	pair, _ := svc.GenerateTokenPair("u1", "e@x.com")

	newPair, err := svc.RefreshAccessToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}
	if newPair.AccessToken == "" {
		t.Error("expected a new access token")
	}
}

func TestRevokeTokenWithoutRedisFails(t *testing.T) {
	svc := newTestJWTService()
	pair, _ := svc.GenerateTokenPair("u1", "e@x.com")

	err := svc.RevokeToken(pair.AccessToken, AccessToken)
	if err == nil {
		t.Fatal("expected revocation to fail without Redis backing store")
	}
}

func TestGetClaimsFromToken(t *testing.T) {
	svc := newTestJWTService()
	pair, _ := svc.GenerateTokenPair("u1", "e@x.com")

	claims, err := svc.GetClaimsFromToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("GetClaimsFromToken failed: %v", err)
	}
	if claims.TokenID == "" {
		t.Error("expected token id claim")
	}
}
