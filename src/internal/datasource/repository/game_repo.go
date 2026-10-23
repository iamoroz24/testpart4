package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"project03/internal/datasource/dbmodels"
	"project03/internal/datasource/mappers"
	"project03/internal/domain/models"
	"project03/internal/domain/ports"
)

type gameRepository struct {
	pool *pgxpool.Pool
}

var _ ports.GameRepository = (*gameRepository)(nil)

func NewGameRepository(pool *pgxpool.Pool) ports.GameRepository {
	return &gameRepository{pool: pool}
}

func (r *gameRepository) Save(ctx context.Context, game models.Game) error {
	dto := mappers.ToDTO(game)
	if dto.ID == uuid.Nil {
		return errors.New("invalid game ID")
	}

	board, err := json.Marshal(dto.Board)
	if err != nil {
		return fmt.Errorf("marshal game board: %w", err)
	}

	var winner any
	if dto.Winner != nil {
		winner = *dto.Winner
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO games (
			id, board, current_player, status, winner, mode, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id) DO UPDATE SET
			board = EXCLUDED.board,
			current_player = EXCLUDED.current_player,
			status = EXCLUDED.status,
			winner = EXCLUDED.winner,
			mode = EXCLUDED.mode,
			updated_at = NOW()
	`,
		dto.ID,
		board,
		dto.CurrentPlayer,
		dto.Status,
		winner,
		dto.Mode,
	)
	if err != nil {
		return fmt.Errorf("save game: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM game_players WHERE game_id = $1`, dto.ID)
	if err != nil {
		return fmt.Errorf("delete old game players: %w", err)
	}

	for _, player := range dto.Players {
		isComputer := game.Mode == models.GameWithComputer &&
			player.Cell == int(models.O)

		_, err = tx.Exec(ctx, `
			INSERT INTO game_players (game_id, player_id, cell, is_computer)
			VALUES ($1, $2, $3, $4)
		`,
			dto.ID,
			player.ID,
			player.Cell,
			isComputer,
		)
		if err != nil {
			return fmt.Errorf("save game player: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit game transaction: %w", err)
	}

	return nil
}

func (r *gameRepository) Get(ctx context.Context, id string) (models.Game, error) {
	gameID, err := uuid.Parse(id)
	if err != nil {
		return models.Game{}, fmt.Errorf("invalid game ID: %w", err)
	}

	var dto dbmodels.GameDTO
	var boardJSON []byte
	var status int16
	var mode int16
	var winner pgtype.UUID

	err = r.pool.QueryRow(ctx, `
		SELECT id, board, current_player, status, winner, mode
		FROM games
		WHERE id = $1
	`, gameID).Scan(
		&dto.ID,
		&boardJSON,
		&dto.CurrentPlayer,
		&status,
		&winner,
		&mode,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Game{}, errors.New("game not found")
		}
		return models.Game{}, fmt.Errorf("get game: %w", err)
	}

	if err := json.Unmarshal(boardJSON, &dto.Board); err != nil {
		return models.Game{}, fmt.Errorf("unmarshal game board: %w", err)
	}

	dto.Status = int(status)
	dto.Mode = int(mode)

	if winner.Valid {
		id := uuid.UUID(winner.Bytes)
		dto.Winner = &id
	}

	rows, err := r.pool.Query(ctx, `
		SELECT player_id, cell
		FROM game_players
		WHERE game_id = $1
		ORDER BY cell
	`, gameID)
	if err != nil {
		return models.Game{}, fmt.Errorf("get game players: %w", err)
	}
	defer rows.Close()

	dto.Players = make([]dbmodels.PlayerDTO, 0, 2)

	for rows.Next() {
		var player dbmodels.PlayerDTO
		var cell int16

		if err := rows.Scan(&player.ID, &cell); err != nil {
			return models.Game{}, fmt.Errorf("scan game player: %w", err)
		}

		player.Cell = int(cell)
		dto.Players = append(dto.Players, player)
	}

	if err := rows.Err(); err != nil {
		return models.Game{}, fmt.Errorf("iterate game players: %w", err)
	}

	return mappers.ToDomain(dto), nil
}


func (r *gameRepository) GetAvailableGames(ctx context.Context) ([]models.Game, error)  {
	rows, err := r.pool.Query(ctx, `
		SELECT
			g.id,
			g.board,
			g.current_player,
			g.status,
			g.winner,
			g.mode,
			p.player_id,
			p.cell
		FROM games g
		JOIN game_players p ON p.game_id = g.id
		WHERE g.status = $1
		  AND (
			  SELECT COUNT(*)
			  FROM game_players gp
			  WHERE gp.game_id = g.id
		  ) = 1
		ORDER BY g.created_at DESC
	`, int16(models.GameWaiting))
	if err != nil {
		return nil, fmt.Errorf("get available games: %w", err)
	}
	defer rows.Close()

	games := make([]models.Game, 0)

	for rows.Next() {
		var dto dbmodels.GameDTO
		var boardJSON []byte
		var status, mode, cell int16
		var winner pgtype.UUID
		var playerID uuid.UUID

		if err := rows.Scan(
			&dto.ID,
			&boardJSON,
			&dto.CurrentPlayer,
			&status,
			&winner,
			&mode,
			&playerID,
			&cell,
		); err != nil {
			return nil, fmt.Errorf("scan available game: %w", err)
		}

		if err := json.Unmarshal(boardJSON, &dto.Board); err != nil {
			return nil, fmt.Errorf("unmarshal available game board: %w", err)
		}

		dto.Status = int(status)
		dto.Mode = int(mode)

		if winner.Valid {
			id := uuid.UUID(winner.Bytes)
			dto.Winner = &id
		}

		dto.Players = []dbmodels.PlayerDTO{
			{
				ID:   playerID,
				Cell: int(cell),
			},
		}

		games = append(games, mappers.ToDomain(dto))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available games: %w", err)
	}

	return games, nil
}


func (r *gameRepository) Delete(ctx context.Context, id string) error {
	gameID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid game ID: %w", err)
	}

	_, err = r.pool.Exec(ctx, `DELETE FROM games WHERE id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("delete game: %w", err)
	}

	return nil
}
