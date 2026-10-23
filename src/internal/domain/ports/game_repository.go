package ports

import (
	"context"

	"project03/internal/domain/models"
)

type GameRepository interface {
	Save(ctx context.Context, game models.Game) error
	Get(ctx context.Context, id string) (models.Game, error)
	GetAvailableGames(ctx context.Context) ([]models.Game, error)
	Delete(ctx context.Context, id string) error
}
