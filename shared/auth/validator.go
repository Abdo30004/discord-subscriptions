package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnauthorized = errors.New("unauthorized: missing or invalid authorization token")
	ErrForbidden    = errors.New("forbidden: administrator privileges required")
)

type UserClaims struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Avatar       string `json:"avatar"`
	GlobalName   string `json:"global_name,omitempty"`
	IsAdmin      bool   `json:"is_admin"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

type Validator struct {
	secretKey []byte
}

func NewValidator(secretKey string) *Validator {
	return &Validator{
		secretKey: []byte(secretKey),
	}
}

func (v *Validator) ValidateToken(tokenString string) (*UserClaims, error) {
	if v == nil || len(v.secretKey) == 0 {
		return nil, errors.New("validator not initialized with secret key")
	}

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, ErrUnauthorized
	}
	return claims, nil
}

func (v *Validator) ExtractAndValidate(r *http.Request) (*UserClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return v.ValidateToken(strings.TrimSpace(authHeader[7:]))
	}

	// Also support reading from HttpOnly cookie
	if cookie, err := r.Cookie("auth_token"); err == nil && cookie.Value != "" {
		return v.ValidateToken(cookie.Value)
	}

	return nil, ErrUnauthorized
}
