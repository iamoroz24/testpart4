package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"project03/internal/domain/models"
	"project03/internal/domain/ports"
	"project03/internal/domain/strategy"
)

type gameService struct {
	repo     ports.GameRepository
	strategy strategy.MoveStrategy
}

func NewService(
	repo ports.GameRepository,
	strategy strategy.MoveStrategy,
) GameService {
	return &gameService{
		repo:     repo,
		strategy: strategy,
	}
}

func (s *gameService) CreateGame(
	ctx context.Context,
	mode models.GameMode,
	playerID uuid.UUID,
) (models.Game, error) {
	game := models.NewGame()
	game.Mode = mode

	switch mode {
	case models.GameWithPlayer:
		if err := game.JoinPlayer(playerID, models.X); err != nil {
			return models.Game{}, err
		}

	case models.GameWithComputer:
		if err := game.JoinPlayer(playerID, models.X); err != nil {
			return models.Game{}, err
		}

		computerID := uuid.New()

		if err := game.JoinPlayer(computerID, models.O); err != nil {
			return models.Game{}, err
		}
		if game.Status != models.GameInProgress {
			return models.Game{}, errors.New("computer game did not start")
		}

	default:
		return models.Game{}, errors.New("invalid game mode")
	}

	if err := s.repo.Save(ctx, game); err != nil {
		return models.Game{}, err
	}

	return game, nil
}
func (s *gameService) GetGame(ctx context.Context, id string) (models.Game, error) {
	return s.repo.Get(ctx, id)
}

func (s *gameService) JoinGame(
	ctx context.Context,
	gameID string,
	playerID uuid.UUID,
	cell models.Cell,
) error {
	game, err := s.repo.Get(ctx, gameID)
	if err != nil {
		return err
	}

	if err := game.JoinPlayer(playerID, cell); err != nil {
		return err
	}

	return s.repo.Save(ctx, game)
}

func (s *gameService) MakeMove(
	ctx context.Context,
	gameID string,
	playerID uuid.UUID,
	row, col int,
) error {
	game, err := s.repo.Get(ctx, gameID)
	if err != nil {
		return err
	}

	// Ход пользователя.
	if err := game.MakeMove(playerID, row, col); err != nil {
		return err
	}

	// Ответ компьютера.
	if game.Mode == models.GameWithComputer &&
		game.Status == models.GameInProgress {

		var computer *models.Player

		for i := range game.Players {
			if game.Players[i].ID != playerID {
				computer = &game.Players[i]
				break
			}
		}

		if computer == nil {
			return errors.New("computer player not found")
		}

		computerRow, computerCol, err := s.strategy.NextMove(
			game.Board,
			computer.Cell,
		)
		if err != nil {
			return err
		}

		if err := game.MakeMove(
			computer.ID,
			computerRow,
			computerCol,
		); err != nil {
			return err
		}
	}

	return s.repo.Save(ctx, game)
}

func (s *gameService) GetAvailableGames(ctx context.Context) ([]models.Game, error) {
	return s.repo.GetAvailableGames(ctx)
}