package strategy

import (
	"testing"

	"project03/internal/domain/models"
)

func TestMinimaxStrategy(t *testing.T) {
	strategy := NewMinimaxStrategy()

	board := models.Board{
		{models.X, models.X, models.Empty},
		{models.Empty, models.O, models.Empty},
		{models.Empty, models.Empty, models.O},
	}

	row, col, err := strategy.NextMove(board, models.O)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if row != 0 || col != 2 {
		t.Fatalf("expected move (0, 2), got (%d, %d)", row, col)
	}
}
