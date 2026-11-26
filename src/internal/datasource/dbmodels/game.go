package dbmodels

import (
	"time"

	"github.com/google/uuid"
)

type PlayerDTO struct {
	ID         uuid.UUID `db:"player_id"`
	Cell       int       `db:"cell"`
	IsComputer bool      `db:"is_computer"`
}

type GameDTO struct {
	ID            uuid.UUID   `db:"id"`
	Board         [3][3]int   `db:"board"`
	Players       []PlayerDTO `db:"players"`
	CurrentPlayer uuid.UUID   `db:"current_player"`
	Status        int         `db:"status"`
	Winner        *uuid.UUID  `db:"winner"`
	Mode          int         `db:"mode"`
	CreatedAt     time.Time   `db:"created_at"`
}
