package middleware

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	UserClass string `json:"class"`
	jwt.RegisteredClaims
}
