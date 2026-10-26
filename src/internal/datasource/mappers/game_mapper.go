package mappers

import (
	"github.com/google/uuid"

	"project03/internal/datasource/dbmodels"
	"project03/internal/domain/models"
)

func ToDomain(dto dbmodels.GameDTO) models.Game {
	var board models.Board

	for i := range board {
		for j := range board[i] {
			board[i][j] = models.Cell(dto.Board[i][j])
		}
	}

	players := make([]models.Player, 0, len(dto.Players))
	for _, player := range dto.Players {
		players = append(players, models.Player{
			ID:         player.ID,
			Cell:       models.Cell(player.Cell),
			IsComputer: player.IsComputer,
		})
	}

	var winner uuid.UUID
	if dto.Winner != nil {
		winner = *dto.Winner
	}

	return models.Game{
		ID:            dto.ID.String(),
		Board:         board,
		Players:       players,
		CurrentPlayer: dto.CurrentPlayer,
		Status:        models.GameStatus(dto.Status),
		Winner:        winner,
		Mode:          models.GameMode(dto.Mode),
	}
}

func ToDTO(game models.Game) dbmodels.GameDTO {
	var board [3][3]int

	for i := range game.Board {
		for j := range game.Board[i] {
			board[i][j] = int(game.Board[i][j])
		}
	}

	players := make([]dbmodels.PlayerDTO, 0, len(game.Players))
	for _, player := range game.Players {
		players = append(players, dbmodels.PlayerDTO{
			ID:         player.ID,
			Cell:       int(player.Cell),
			IsComputer: player.IsComputer,
		})
	}

	id, err := uuid.Parse(game.ID)
	if err != nil {
		id = uuid.Nil
	}

	dto := dbmodels.GameDTO{
		ID:            id,
		Board:         board,
		Players:       players,
		CurrentPlayer: game.CurrentPlayer,
		Status:        int(game.Status),
		Mode:          int(game.Mode),
	}

	if game.Winner != uuid.Nil {
		winner := game.Winner
		dto.Winner = &winner
	}

	return dto
}
