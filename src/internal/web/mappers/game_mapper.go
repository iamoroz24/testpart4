package mappers

import (
	"project03/internal/domain/models"
	"project03/internal/web/webmodels"
)

func ToWebResponse(game models.Game, winner string) webmodels.GameResponse {
	var board [3][3]int

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			board[i][j] = int(game.Board[i][j])
		}
	}

	return webmodels.GameResponse{
		ID:     game.ID,
		Board:  board,
		Winner: winner,
	}
}

func ToNewGameResponse(game models.Game) webmodels.NewGameResponse {
	var board [3][3]int

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			board[i][j] = int(game.Board[i][j])
		}
	}

	return webmodels.NewGameResponse{
		ID:    game.ID,
		Board: board,
		Turn:  0,
	}
}
