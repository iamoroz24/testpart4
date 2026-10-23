package service

import (
	"context"

	"project03/internal/domain/models"
)

type UserService interface {
	Register(ctx context.Context, login, password string) (models.User, error)
	Authenticate(ctx context.Context, login, password string) (models.User, error)
	GetByID(ctx context.Context, id string) (models.User, error)
}

/*Метод Authenticate будет
находить пользовател по логину
проверять пароль с помощью bcrypt
возвращать пользователя при успешной проверке
возвращать ошибку при неверном логине или пароле
*/