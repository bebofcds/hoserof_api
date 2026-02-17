/*
================================================================================
HOSEROF_API - JWT Service
================================================================================

Description:
This package provides JWT token generation for authentication and authorization
within the Hoserof system. It leverages the `github.com/golang-jwt/jwt/v5`
library and custom middleware claims to issue signed tokens for students and staff.

Responsibilities:
1. Init(secret []byte) *JWTService
   - Initializes a JWT service with a secret key for signing tokens.

2. GenerateToken(id, class, role, name string) (string, error)
   - Generates a JWT token with the following claims:
     - ID         : Unique identifier of the user (student or staff)
     - UserClass  : Class identifier (if applicable)
     - Name       : User's display name
     - Role       : User role (e.g., student, teacher, admin)
     - IssuedAt   : Token issue time
     - NotBefore  : Token valid from this time
     - ExpiresAt  : Token expiration (6 months from issuance)
     - Issuer     : "hoserof"

Usage Notes:
- The token is signed with HS256 using the secret provided during initialization.
- Tokens should be passed in the `Authorization` header as `Bearer <token>` for protected routes.
- The Claims structure is defined in `middleware.Claims` to include custom fields.

Security Notes:
- Keep the `Secret` value confidential and rotate periodically.
- Expiration is set to 6 months; adjust as needed for security policies.
- Ensure all protected endpoints verify tokens using the same secret.

Example Usage:

	jwtService := Init([]byte(os.Getenv("JWT_KEY")))
	token, err := jwtService.GenerateToken(userID, userClass, role, name)
	if err != nil {
	    log.Fatalf("Failed to generate JWT: %v", err)
	}
================================================================================
*/

package services

import (
	"HOSEROF_API/middleware"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Secret []byte
}

func Init(secret []byte) *JWTService {
	return &JWTService{
		Secret: secret,
	}
}

func (j *JWTService) GenerateToken(id, class, role, name string) (string, error) {

	claims := middleware.Claims{
		ID:        id,
		UserClass: class,
		Name:      name,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 6, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "hoserof",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.Secret)

	return signed, err
}
