package service

import (
	"context"

	"github.com/google/uuid"

	"project03/internal/domain/models"
)

type GameService interface {
	CreateGame(ctx context.Context, mode models.GameMode, playerID uuid.UUID) (models.Game, error)
	GetGame(ctx context.Context, id string) (models.Game, error)
	JoinGame(ctx context.Context, gameID string, playerID uuid.UUID, cell models.Cell) error
	MakeMove(ctx context.Context, gameID string, playerID uuid.UUID, row, col int) error
	GetAvailableGames(ctx context.Context) ([]models.Game, error)
}
