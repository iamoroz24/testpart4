package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"project03/internal/domain/models"
	"project03/internal/domain/ports"
)

var (
	ErrInvalidLogin       = errors.New("login must not be empty")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type userService struct {
	repo ports.UserRepository
}

var _ UserService = (*userService)(nil)

func NewUserService(repo ports.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(
	ctx context.Context,
	login, password string,
) (models.User, error) {
	login = strings.TrimSpace(login)

	if login == "" {
		return models.User{}, ErrInvalidLogin
	}

	if len(password) < 8 {
		return models.User{}, ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: string(hash),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return models.User{}, err
	}

	// Не возвращаем хеш пароля вызывающему коду.
	user.PasswordHash = ""

	return user, nil
}

func (s *userService) GetByID(
	ctx context.Context,
	id string,
) (models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) Authenticate(
	ctx context.Context,
	login string,
	password string,
) (models.User, error) {
	login = strings.TrimSpace(login)

	if login == "" || password == "" {
		return models.User{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.User{}, ErrInvalidCredentials
		}

		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return models.User{}, ErrInvalidCredentials
	}

	user.PasswordHash = ""
	return user, nil
}
