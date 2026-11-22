package jwt

import (
	"fmt"
	"project03/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtProvider struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

const TokenTypeAccess = "access"
const TokenTypeRefresh = "refresh"

func NewJwtProvider(secretKey []byte) (*JwtProvider, error) {

	if len(secretKey) < 32 {
		return nil, fmt.Errorf("secret key must be at least 32 bytes")
	}

	return &JwtProvider{
		secretKey:       secretKey,
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}, nil

}

func (j *JwtProvider) GenerateAccessToken(user models.User) (string, error) {

	now := time.Now()

	claims := &Claims{
		UserID:    user.ID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secretKey)

}

func (j *JwtProvider) GenerateRefreshToken(user models.User) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:    user.ID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JwtProvider) parseAndValidate(tokenStr, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return j.secretKey, nil
		})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: %s", claims.TokenType)
	}
	if claims.UserID == uuid.Nil {
		return nil, fmt.Errorf("missing user ID")
	}

	return claims, nil
}

func (j *JwtProvider) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return j.parseAndValidate(tokenStr, TokenTypeAccess)
}

func (j *JwtProvider) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return j.parseAndValidate(tokenStr, TokenTypeRefresh)
}

func (j *JwtProvider) GetUserIDFromToken(tokenStr string) (uuid.UUID, error) {
	claims, err := j.ValidateAccessToken(tokenStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("validate access token: %w", err)
	}
	return claims.UserID, nil
}
