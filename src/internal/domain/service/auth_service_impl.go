package service

import (
	"context"
	"errors"
	"fmt"
)

type authService struct {
	userService   UserService
	tokenProvider TokenProvider
}

func NewAuthService(userService UserService, tokenProvider TokenProvider) AuthService {
	return &authService{
		userService:   userService,
		tokenProvider: tokenProvider,
	}
}

var (
	ErrInvalidToken = errors.New("invalid token")
)

func (c *authService) Login(ctx context.Context, login, password string) (TokenPair, error) {

	user, err := c.userService.Authenticate(ctx, login, password)

	if err != nil {
		return TokenPair{}, err
	}

	accessToken, err := c.tokenProvider.GenerateAccessToken(user)

	if err != nil {
		return TokenPair{}, fmt.Errorf("generate acces token: %w", err)
	}

	refreshToken, err := c.tokenProvider.GenerateRefreshToken(user)

	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func (c *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := c.tokenProvider.ValidateRefreshToken(refreshToken)
	if err != nil {
		return TokenPair{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	user, err := c.userService.GetByID(ctx, claims.UserID.String())
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return TokenPair{}, ErrInvalidToken
		}

		return TokenPair{}, fmt.Errorf("get user by id: %w", err)
	}

	accessToken, err := c.tokenProvider.GenerateAccessToken(user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate access token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (c *authService) RefreshRefreshToken(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := c.tokenProvider.ValidateRefreshToken(refreshToken)
	if err != nil {
		return TokenPair{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	user, err := c.userService.GetByID(ctx, claims.UserID.String())
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return TokenPair{}, ErrInvalidToken
		}

		return TokenPair{}, fmt.Errorf("get user by id: %w", err)
	}

	accessToken, err := c.tokenProvider.GenerateAccessToken(user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := c.tokenProvider.GenerateRefreshToken(user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
