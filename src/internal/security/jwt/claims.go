package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType string    `json:"token_type"`

	jwt.RegisteredClaims
}
