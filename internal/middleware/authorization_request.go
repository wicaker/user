package middleware

import (
	"os"
	"strings"

	"github.com/wicaker/user/internal/domain"

	jwt "github.com/dgrijalva/jwt-go"
)

// JwtVerify will validate and parsing an incoming jwt token
func JwtVerify(token string) (*domain.JWToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		//Token is missing, returns with error code 403 Unauthorized
		return nil, domain.ErrUnauthorized
	}

	parsedToken := &domain.JWToken{}
	_, err := jwt.ParseWithClaims(token, parsedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return parsedToken, nil
}
