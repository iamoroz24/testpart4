package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"project03/internal/domain/models"
	"project03/internal/security/jwt"
)

// ---------- моки ----------

type mockUserService struct {
	authenticateFn func(ctx context.Context, login, password string) (models.User, error)
	getByIDFn      func(ctx context.Context, id string) (models.User, error)
	registerFn     func(ctx context.Context, login, password string) (models.User, error)

	authenticateCalled bool
	getByIDCalled      bool
	lastGetByIDArg     string
}

func (m *mockUserService) Authenticate(ctx context.Context, login, password string) (models.User, error) {
	m.authenticateCalled = true
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, login, password)
	}
	return models.User{}, errors.New("authenticateFn not set")
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (models.User, error) {
	m.getByIDCalled = true
	m.lastGetByIDArg = id
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return models.User{}, errors.New("getByIDFn not set")
}

func (m *mockUserService) Register(ctx context.Context, login, password string) (models.User, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, login, password)
	}
	return models.User{}, errors.New("register not implemented in mock")
}

type mockTokenProvider struct {
	generateAccessFn  func(user models.User) (string, error)
	generateRefreshFn func(user models.User) (string, error)
	validateRefreshFn func(token string) (*jwt.Claims, error)

	generateAccessCalled  bool
	generateRefreshCalled bool
	validateRefreshCalled bool
	lastValidateArg       string
}

func (m *mockTokenProvider) GenerateAccessToken(user models.User) (string, error) {
	m.generateAccessCalled = true
	if m.generateAccessFn != nil {
		return m.generateAccessFn(user)
	}
	return "", errors.New("generateAccessFn not set")
}

func (m *mockTokenProvider) GenerateRefreshToken(user models.User) (string, error) {
	m.generateRefreshCalled = true
	if m.generateRefreshFn != nil {
		return m.generateRefreshFn(user)
	}
	return "", errors.New("generateRefreshFn not set")
}

func (m *mockTokenProvider) ValidateRefreshToken(token string) (*jwt.Claims, error) {
	m.validateRefreshCalled = true
	m.lastValidateArg = token
	if m.validateRefreshFn != nil {
		return m.validateRefreshFn(token)
	}
	return nil, errors.New("validateRefreshFn not set")
}

// Заглушки, если интерфейс TokenProvider требует больше методов
func (m *mockTokenProvider) ValidateAccessToken(token string) (*jwt.Claims, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockTokenProvider) GetUserIDFromToken(token string) (uuid.UUID, error) {
	return uuid.Nil, errors.New("not implemented in mock")
}

// ---------- Login ----------

func TestAuthService_Login_Success(t *testing.T) {
	user := models.User{ID: uuid.New(), Login: "alice"}

	userService := &mockUserService{
		authenticateFn: func(ctx context.Context, login, password string) (models.User, error) {
			return user, nil
		},
	}
	tokenProvider := &mockTokenProvider{
		generateAccessFn:  func(u models.User) (string, error) { return "access-token", nil },
		generateRefreshFn: func(u models.User) (string, error) { return "refresh-token", nil },
	}

	svc := NewAuthService(userService, tokenProvider)

	pair, err := svc.Login(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.AccessToken != "access-token" {
		t.Errorf("expected access-token, got %q", pair.AccessToken)
	}
	if pair.RefreshToken != "refresh-token" {
		t.Errorf("expected refresh-token, got %q", pair.RefreshToken)
	}
	if !userService.authenticateCalled {
		t.Error("expected Authenticate to be called")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userService := &mockUserService{
		authenticateFn: func(ctx context.Context, login, password string) (models.User, error) {
			return models.User{}, ErrInvalidCredentials
		},
	}
	tokenProvider := &mockTokenProvider{}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.Login(context.Background(), "alice", "wrong")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
	if tokenProvider.generateAccessCalled {
		t.Error("GenerateAccessToken should not be called on auth failure")
	}
	if tokenProvider.generateRefreshCalled {
		t.Error("GenerateRefreshToken should not be called on auth failure")
	}
}

// ---------- RefreshAccessToken ----------

func TestAuthService_RefreshAccessToken_Success(t *testing.T) {
	user := models.User{ID: uuid.New(), Login: "alice"}

	userService := &mockUserService{
		getByIDFn: func(ctx context.Context, id string) (models.User, error) {
			if id != user.ID.String() {
				t.Errorf("expected id %q, got %q", user.ID.String(), id)
			}
			return user, nil
		},
	}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: user.ID}, nil
		},
		generateAccessFn:  func(u models.User) (string, error) { return "new-access", nil },
		generateRefreshFn: func(u models.User) (string, error) { return "new-refresh", nil },
	}

	svc := NewAuthService(userService, tokenProvider)

	pair, err := svc.RefreshAccessToken(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.AccessToken != "new-access" {
		t.Errorf("expected new-access, got %q", pair.AccessToken)
	}
	if pair.RefreshToken != "old-refresh" {
		t.Errorf("RefreshAccessToken должен сохранять старый refresh, got %q", pair.RefreshToken)
	}
	if tokenProvider.generateRefreshCalled {
		t.Error("GenerateRefreshToken не должен вызываться в RefreshAccessToken")
	}
	if !tokenProvider.validateRefreshCalled {
		t.Error("expected ValidateRefreshToken to be called")
	}
}

func TestAuthService_RefreshAccessToken_InvalidToken(t *testing.T) {
	userService := &mockUserService{}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return nil, errors.New("jwt expired")
		},
	}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.RefreshAccessToken(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
	if userService.getByIDCalled {
		t.Error("GetByID не должен вызываться при невалидном токене")
	}
}

func TestAuthService_RefreshAccessToken_UserNotFound(t *testing.T) {
	userService := &mockUserService{
		getByIDFn: func(ctx context.Context, id string) (models.User, error) {
			return models.User{}, ErrUserNotFound
		},
	}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: uuid.New()}, nil
		},
	}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.RefreshAccessToken(context.Background(), "valid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ErrUserNotFound должен превращаться в ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_GenerateError(t *testing.T) {
	user := models.User{ID: uuid.New()}
	userService := &mockUserService{
		getByIDFn: func(ctx context.Context, id string) (models.User, error) { return user, nil },
	}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: user.ID}, nil
		},
		generateAccessFn: func(u models.User) (string, error) {
			return "", errors.New("signing failed")
		},
	}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.RefreshAccessToken(context.Background(), "valid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------- RefreshRefreshToken ----------

func TestAuthService_RefreshRefreshToken_Success(t *testing.T) {
	user := models.User{ID: uuid.New(), Login: "alice"}

	userService := &mockUserService{
		getByIDFn: func(ctx context.Context, id string) (models.User, error) {
			return user, nil
		},
	}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: user.ID}, nil
		},
		generateAccessFn:  func(u models.User) (string, error) { return "new-access", nil },
		generateRefreshFn: func(u models.User) (string, error) { return "new-refresh", nil },
	}

	svc := NewAuthService(userService, tokenProvider)

	pair, err := svc.RefreshRefreshToken(context.Background(), "old-refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair.AccessToken != "new-access" {
		t.Errorf("expected new-access, got %q", pair.AccessToken)
	}
	if pair.RefreshToken != "new-refresh" {
		t.Errorf("RefreshRefreshToken должен выдавать новый refresh, got %q", pair.RefreshToken)
	}
	if !tokenProvider.generateRefreshCalled {
		t.Error("GenerateRefreshToken должен вызываться в RefreshRefreshToken")
	}
}

func TestAuthService_RefreshRefreshToken_InvalidToken(t *testing.T) {
	userService := &mockUserService{}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return nil, errors.New("bad token")
		},
	}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.RefreshRefreshToken(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshRefreshToken_UserNotFound(t *testing.T) {
	userService := &mockUserService{
		getByIDFn: func(ctx context.Context, id string) (models.User, error) {
			return models.User{}, ErrUserNotFound
		},
	}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: uuid.New()}, nil
		},
	}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.RefreshRefreshToken(context.Background(), "valid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("ErrUserNotFound должен превращаться в ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshRefreshToken_GenerateRefreshError(t *testing.T) {
	user := models.User{ID: uuid.New()}
	userService := &mockUserService{
		getByIDFn: func(ctx context.Context, id string) (models.User, error) { return user, nil },
	}
	tokenProvider := &mockTokenProvider{
		validateRefreshFn: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: user.ID}, nil
		},
		generateAccessFn: func(u models.User) (string, error) { return "access", nil },
		generateRefreshFn: func(u models.User) (string, error) {
			return "", errors.New("signing failed")
		},
	}

	svc := NewAuthService(userService, tokenProvider)

	_, err := svc.RefreshRefreshToken(context.Background(), "valid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
