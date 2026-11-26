package service

import (
	"context"
	"project03/internal/domain/models"
	"project03/internal/security/jwt"
)

type AuthService interface {
	Login(ctx context.Context, login, password string) (TokenPair, error)

	RefreshAccessToken(ctx context.Context, refreshToken string) (TokenPair, error)

	RefreshRefreshToken(ctx context.Context, refreshToken string) (TokenPair, error)
}

type TokenProvider interface {
	GenerateAccessToken(user models.User) (string, error)
	GenerateRefreshToken(user models.User) (string, error)
	ValidateRefreshToken(token string) (*jwt.Claims, error)
	//GetUserIDFromToken(token string) (uuid.UUID, error)
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
