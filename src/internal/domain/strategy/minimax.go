package strategy

import (
	"errors"

	"project03/internal/domain/models"
)

type MinimaxStrategy struct{}

func NewMinimaxStrategy() MoveStrategy {
	return &MinimaxStrategy{}
}

func (s *MinimaxStrategy) NextMove(
	board models.Board,
	cell models.Cell,
) (int, int, error) {
	game := models.Game{
		Board: board,
	}

	if game.Status == models.GameWon || game.Status == models.GameDraw {
		return 0, 0, errors.New("game is already over")
	}

	opponent := models.X
	if cell == models.X {
		opponent = models.O
	}

	bestScore := -1000
	bestRow := -1
	bestCol := -1

	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if board[row][col] != models.Empty {
				continue
			}

			board[row][col] = cell

			score := s.minimax(board, cell, opponent, 0, false)

			board[row][col] = models.Empty

			if score > bestScore {
				bestScore = score
				bestRow = row
				bestCol = col
			}
		}
	}

	if bestRow == -1 {
		return 0, 0, errors.New("no available moves")
	}

	return bestRow, bestCol, nil
}

func checkWinner(board models.Board) (models.Cell, bool) {
	for i := 0; i < 3; i++ {
		if board[i][0] != models.Empty &&
			board[i][0] == board[i][1] &&
			board[i][1] == board[i][2] {
			return board[i][0], true
		}

		if board[0][i] != models.Empty &&
			board[0][i] == board[1][i] &&
			board[1][i] == board[2][i] {
			return board[0][i], true
		}
	}

	if board[0][0] != models.Empty &&
		board[0][0] == board[1][1] &&
		board[1][1] == board[2][2] {
		return board[0][0], true
	}

	if board[0][2] != models.Empty &&
		board[0][2] == board[1][1] &&
		board[1][1] == board[2][0] {
		return board[0][2], true
	}

	return models.Empty, false
}


func (s *MinimaxStrategy) minimax(
	board models.Board,
	maximizingCell models.Cell,
	opponentCell models.Cell,
	depth int,
	isMaximizing bool,
) int {
	winner, ok := checkWinner(board)

	if ok {
		if winner == maximizingCell {
			return 10 - depth
		}

		return -10 + depth
	}

	if board.IsFull() {
		return 0
	}

	if isMaximizing {
		bestScore := -1000

		for row := 0; row < 3; row++ {
			for col := 0; col < 3; col++ {
				if board[row][col] != models.Empty {
					continue
				}

				board[row][col] = maximizingCell

				score := s.minimax(
					board,
					maximizingCell,
					opponentCell,
					depth+1,
					false,
				)

				board[row][col] = models.Empty

				bestScore = max(bestScore, score)
			}
		}

		return bestScore
	}

	bestScore := 1000

	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			if board[row][col] != models.Empty {
				continue
			}

			board[row][col] = opponentCell

			score := s.minimax(
				board,
				maximizingCell,
				opponentCell,
				depth+1,
				true,
			)

			board[row][col] = models.Empty

			bestScore = min(bestScore, score)
		}
	}

	return bestScore
}

