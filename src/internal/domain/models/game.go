package models

import (
	"errors"

	"github.com/google/uuid"
)

type GameStatus int

const (
	GameWaiting GameStatus = iota
	GameInProgress
	GameDraw
	GameWon
)

type Player struct {
	ID   uuid.UUID
	Cell Cell
	IsComputer bool
}

type GameMode int

const (
	GameWithPlayer GameMode = iota
	GameWithComputer
)

type Game struct {
	ID            string
	Board         Board
	Players       []Player
	CurrentPlayer uuid.UUID
	Status        GameStatus
	Winner        uuid.UUID
	Mode          GameMode
}

func NewGame() Game {
	return Game{
		ID:      uuid.New().String(),
		Board:   NewBoard(),
		Players: []Player{},
		Status:  GameWaiting,
		Winner:  uuid.Nil,
		Mode:    GameWithPlayer,
	}
}

var (
	ErrGameFinished        = errors.New("game is already finished")
	ErrGameNotStarted      = errors.New("game has not started")
	ErrPlayerNotFound      = errors.New("player is not in game")
	ErrNotPlayerTurn       = errors.New("not player's turn")
	ErrInvalidPosition     = errors.New("invalid board position")
	ErrCellOccupied        = errors.New("cell is already occupied")
	ErrPlayerAlreadyJoined = errors.New("player already joined the game")
	ErrGameFull            = errors.New("game already has two players")
	ErrInvalidPlayerCell   = errors.New("invalid player cell")
)

func (g Game) CheckWinner() (Cell, bool) {
	board := g.Board

	for i := 0; i < 3; i++ {
		if board[i][0] != Empty &&
			board[i][0] == board[i][1] &&
			board[i][1] == board[i][2] {
			return board[i][0], true
		}
	}

	for j := 0; j < 3; j++ {
		if board[0][j] != Empty &&
			board[0][j] == board[1][j] &&
			board[1][j] == board[2][j] {
			return board[0][j], true
		}
	}

	if board[0][0] != Empty &&
		board[0][0] == board[1][1] &&
		board[1][1] == board[2][2] {
		return board[0][0], true
	}

	if board[0][2] != Empty &&
		board[0][2] == board[1][1] &&
		board[1][1] == board[2][0] {
		return board[0][2], true
	}
	return Empty, false
}

func (g *Game) MakeMove(playerID uuid.UUID, row, col int) error {
	if g.Status == GameWon || g.Status == GameDraw {
		return ErrGameFinished
	}

	if g.Status != GameInProgress {
		return ErrGameNotStarted
	}

	if row < 0 || row >= 3 || col < 0 || col >= 3 {
		return ErrInvalidPosition
	}

	var player *Player

	for i := range g.Players {
		if g.Players[i].ID == playerID {
			player = &g.Players[i]
			break
		}
	}

	if player == nil {
		return ErrPlayerNotFound
	}

	if g.CurrentPlayer != playerID {
		return ErrNotPlayerTurn
	}

	if !g.Board.MakeMove(row, col, player.Cell) {
		return ErrCellOccupied
	}

	if _, ok := g.CheckWinner(); ok {
		g.Status = GameWon
		g.Winner = playerID
		return nil
	}

	if g.Board.IsFull() {
		g.Status = GameDraw
		g.Winner = uuid.Nil
		return nil
	}

	for _, p := range g.Players {
		if p.ID != playerID {
			g.CurrentPlayer = p.ID
			break
		}
	}

	return nil
}

func (g *Game) JoinPlayer(playerID uuid.UUID, cell Cell) error {
	if len(g.Players) >= 2 {
		return ErrGameFull
	}

	if g.Status != GameWaiting {
		return ErrGameNotStarted
	}

	if cell != X && cell != O {
		return ErrInvalidPlayerCell
	}

	for _, player := range g.Players {
		if player.ID == playerID {
			return ErrPlayerAlreadyJoined
		}

		if player.Cell == cell {
			return ErrInvalidPlayerCell
		}
	}

	g.Players = append(g.Players, Player{
		ID:   playerID,
		Cell: cell,
	})

	if len(g.Players) == 1 {
		g.CurrentPlayer = playerID
	}

	if len(g.Players) == 2 {
		g.Status = GameInProgress
	}

	return nil
}
