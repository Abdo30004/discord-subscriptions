package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidator(t *testing.T) {
	secret := "test-secret-key-1234567890123456"
	v := NewValidator(secret)

	t.Run("valid token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
			UserID:       "12345",
			Username:     "testuser",
			IsAdmin:      true,
			IsSuperAdmin: false,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		})
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		claims, err := v.ValidateToken(signed)
		if err != nil {
			t.Fatalf("expected valid token, got: %v", err)
		}
		if claims.UserID != "12345" || !claims.IsAdmin {
			t.Errorf("unexpected claims: %+v", claims)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
			UserID: "12345",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
		})
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}

		_, err = v.ValidateToken(signed)
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})

	t.Run("ExtractAndValidate from Bearer header", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
			UserID: "user-bearer",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		})
		signed, _ := token.SignedString([]byte(secret))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+signed)

		claims, err := v.ExtractAndValidate(req)
		if err != nil {
			t.Fatalf("expected success, got: %v", err)
		}
		if claims.UserID != "user-bearer" {
			t.Errorf("expected user-bearer, got: %s", claims.UserID)
		}
	})

	t.Run("ExtractAndValidate from cookie", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
			UserID: "user-cookie",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		})
		signed, _ := token.SignedString([]byte(secret))

		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: signed,
		})

		claims, err := v.ExtractAndValidate(req)
		if err != nil {
			t.Fatalf("expected success, got: %v", err)
		}
		if claims.UserID != "user-cookie" {
			t.Errorf("expected user-cookie, got: %s", claims.UserID)
		}
	})

	t.Run("ExtractAndValidate unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		_, err := v.ExtractAndValidate(req)
		if err != ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})
}
