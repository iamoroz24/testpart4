package jwt

import (
	"testing"
	"time"

	"project03/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func newTestUser() models.User {
	return models.User{
		ID: uuid.New(),
		Login: "testuser",
	}
}

const testSecret = "01234567890123456789012345678901" // ровно 32 байта


func TestJwtProvider(t *testing.T) {
	t.Run("valid secret", func(t *testing.T) {
		provider, err := NewJwtProvider([]byte(testSecret))

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if provider == nil {
			t.Fatal("expected provider, got nil")
		}
	})


	t.Run("secret to short", func(t *testing.T) {
		_, err := NewJwtProvider([]byte("short"))
		if err == nil {
			t.Fatal("expected error for short secret, got nil")
		}
	})

	t.Run("secret empty", func(t *testing.T) {
		_, err := NewJwtProvider([]byte(""))

		if err == nil {
			t.Fatal("expected error for empty secret, got nil")
		}
	})
}


func TestGenerateAccessToken(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	user := newTestUser()

	tokenStr, err := provider.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	// Проверяем, что токен валиден и содержит правильные claims
	claims, err := provider.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("validate access token: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, claims.UserID)
	}
	if claims.TokenType != TokenTypeAccess {
		t.Errorf("expected token type %q, got %q", TokenTypeAccess, claims.TokenType)
	}
}

func TestValidateToken_WrongAlgorithm(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	user := newTestUser()

	// Генерируем токен с HS512 вручную
	claims := &Claims{
		UserID:    user.ID,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, _ := token.SignedString(provider.secretKey)

	_, err := provider.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for non-HS256 token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	user := newTestUser()

	tokenStr, err := provider.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	claims, err := provider.ValidateRefreshToken(tokenStr)
	if err != nil {
		t.Fatalf("validate refresh token: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, claims.UserID)
	}
	if claims.TokenType != TokenTypeRefresh {
		t.Errorf("expected token type %q, got %q", TokenTypeRefresh, claims.TokenType)
	}
}


func TestTokenTypeMismatch(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	user := newTestUser()

	accessToken, _ := provider.GenerateAccessToken(user)
	refreshToken, _ := provider.GenerateRefreshToken(user)

	t.Run("access token rejected as refresh", func(t *testing.T) {
		_, err := provider.ValidateRefreshToken(accessToken)
		if err == nil {
			t.Fatal("expected error when validating access token as refresh")
		}
	})

	t.Run("refresh token rejected as access", func(t *testing.T) {
		_, err := provider.ValidateAccessToken(refreshToken)
		if err == nil {
			t.Fatal("expected error when validating refresh token as access")
		}
	})
}



func TestValidateToken_WrongSignature(t *testing.T) {
	provider1, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	provider2, _ := NewJwtProvider([]byte("98765432109876543210987654321098"))
	user := newTestUser()

	tokenStr, _ := provider1.GenerateAccessToken(user)

	_, err := provider2.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for token signed with different key")
	}
}


func TestValidateToken_Expired(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	// Подменяем TTL на отрицательный — токен будет просрочен сразу после генерации
	provider.accessTokenTTL = -time.Minute

	user := newTestUser()
	tokenStr, err := provider.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = provider.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}


func TestValidateToken_Garbage(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))

	_, err := provider.ValidateAccessToken("not-a-token")
	if err == nil {
		t.Fatal("expected error for garbage token")
	}

	_, err = provider.ValidateAccessToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}


func TestGetUserIDFromToken(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))
	user := newTestUser()

	tokenStr, _ := provider.GenerateAccessToken(user)

	id, err := provider.GetUserIDFromToken(tokenStr)
	if err != nil {
		t.Fatalf("get user id: %v", err)
	}
	if id != user.ID {
		t.Errorf("expected %v, got %v", user.ID, id)
	}
}

func TestGetUserIDFromToken_Invalid(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))

	_, err := provider.GetUserIDFromToken("garbage")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}


func TestValidateToken_MissingUserID(t *testing.T) {
	provider, _ := NewJwtProvider([]byte("01234567890123456789012345678901"))

	// user.ID == uuid.Nil
	tokenStr, err := provider.GenerateAccessToken(models.User{})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = provider.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for token without user ID")
	}
}