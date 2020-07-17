package domain

import (
	jwt "github.com/dgrijalva/jwt-go"
)

//JWToken struct declaration
type JWToken struct {
	UUID  string
	Email string
	Salt  string
	*jwt.StandardClaims
}
