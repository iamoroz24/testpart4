package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"project03/internal/domain/models"
	"project03/internal/domain/strategy"
)

func setupService() GameService {
	repo := newMockGameRepository()

	return NewService(
		repo,
		strategy.NewMinimaxStrategy(),
	)
}

func (m *mockGameRepository) GetAvailableGames(
	ctx context.Context,
) ([]models.Game, error) {
	return []models.Game{}, nil
}

func TestCreateGame(t *testing.T) {
	service := setupService()

	game, err := service.CreateGame(
		context.Background(),
		models.GameWithPlayer,
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("error creating game: %v", err)
	}

	if game.ID == "" {
		t.Error("expected game ID to be set")
	}

	if game.Status != models.GameWaiting {
		t.Errorf("expected GameWaiting, got %v", game.Status)
	}
}

func TestSaveAndGetGame(t *testing.T) {
	service := setupService()
	ctx := context.Background()

	game, err := service.CreateGame(
		ctx,
		models.GameWithPlayer,
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("error creating game: %v", err)
	}

	loadedGame, err := service.GetGame(ctx, game.ID)
	if err != nil {
		t.Fatalf("error getting game: %v", err)
	}

	if loadedGame.ID != game.ID {
		t.Errorf("expected ID %s, got %s", game.ID, loadedGame.ID)
	}
}

func TestMakeMoveAgainstComputer(t *testing.T) {
	service := setupService()
	ctx := context.Background()

	playerID := uuid.New()

	game, err := service.CreateGame(
		ctx,
		models.GameWithComputer,
		playerID,
	)

	if err != nil {
		t.Fatalf("error creating game: %v", err)
	}

	t.Logf(
		"created game: status=%v players=%d mode=%v current=%s",
		game.Status,
		len(game.Players),
		game.Mode,
		game.CurrentPlayer,
	)

	loadedGame, err := service.GetGame(ctx, game.ID)
	if err != nil {
		t.Fatalf("error getting game: %v", err)
	}

	t.Logf(
		"loaded game: status=%v players=%d mode=%v current=%s",
		loadedGame.Status,
		len(loadedGame.Players),
		loadedGame.Mode,
		loadedGame.CurrentPlayer,
	)
	if err != nil {
		t.Fatalf("error creating game: %v", err)
	}

	if game.Status != models.GameInProgress {
		t.Fatalf("expected game to be in progress, got %v", game.Status)
	}

	if len(game.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(game.Players))
	}

	err = service.MakeMove(
		ctx,
		game.ID,
		playerID,
		0,
		0,
	)
	if err != nil {
		t.Fatalf("error making move: %v", err)
	}

	updatedGame, err := service.GetGame(ctx, game.ID)
	if err != nil {
		t.Fatalf("error getting game: %v", err)
	}

	// Ход пользователя.
	if updatedGame.Board[0][0] != models.X {
		t.Errorf(
			"expected player move at [0][0] to be X, got %v",
			updatedGame.Board[0][0],
		)
	}

	// Компьютер должен был сделать ответный ход.
	moves := 0

	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if updatedGame.Board[row][col] != models.Empty {
				moves++
			}
		}
	}

	if moves != 2 {
		t.Errorf("expected 2 moves after player turn, got %d", moves)
	}
}
