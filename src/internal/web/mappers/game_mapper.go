package mappers

import (
	"project03/internal/domain/models"
	"project03/internal/web/webmodels"
	"github.com/google/uuid"
)

func ToWebResponse(game models.Game, _ string) webmodels.GameResponse {
	return webmodels.GameResponse{
		ID:            game.ID,
		Board:         toWebBoard(game.Board),
		Players:       toWebPlayers(game.Players),
		CurrentPlayer: game.CurrentPlayer.String(),
		Status:        int(game.Status),
		Winner:        winnerID(game.Winner),
		Mode:          int(game.Mode),
	}
}

func ToNewGameResponse(game models.Game) webmodels.NewGameResponse {
	return webmodels.NewGameResponse{
		ID:            game.ID,
		Board:         toWebBoard(game.Board),
		Players:       toWebPlayers(game.Players),
		CurrentPlayer: game.CurrentPlayer.String(),
		Status:        int(game.Status),
		Mode:          int(game.Mode),
	}
}

func toWebBoard(board models.Board) [3][3]int {
	var result [3][3]int

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			result[i][j] = int(board[i][j])
		}
	}

	return result
}

func toWebPlayers(players []models.Player) []webmodels.PlayerResponse {
	result := make([]webmodels.PlayerResponse, 0, len(players))

	for _, player := range players {
		result = append(result, webmodels.PlayerResponse{
			ID:   player.ID.String(),
			Cell: int(player.Cell),
		})
	}

	return result
}


func winnerID(id uuid.UUID) string {
    if id == uuid.Nil {
        return ""
    }

    return id.String()
}