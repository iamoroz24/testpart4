package ports

import (
	"context"

	"project03/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByLogin(ctx context.Context, login string) (models.User, error)
	GetByID(ctx context.Context, id string) (models.User, error)
}

/*Почему интерфейс находится в domain/ports: доменная логика определяет,
какие операции ей нужны, а PostgreSQL — лишь одна из возможных реализаций.*/
