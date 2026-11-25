package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"project03/internal/domain/models"
	"project03/internal/domain/service"
	"project03/internal/web/mappers"
	"project03/internal/web/middleware"
	"project03/internal/web/webmodels"
)

type GameHandler struct {
	service service.GameService
}

func NewGameHandler(service service.GameService) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {

	playerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		sendError(w, "authentication requried", http.StatusUnauthorized)
		return
	}

	var req struct {
		Mode string `json:"mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var mode models.GameMode

	switch req.Mode {
	case "player":
		mode = models.GameWithPlayer
	case "computer":
		mode = models.GameWithComputer
	default:
		sendError(w, "invalid game mode", http.StatusBadRequest)
		return
	}

	game, err := h.service.CreateGame(
		r.Context(),
		mode,
		playerID,
	)
	if err != nil {
		sendError(w, "failed to create game", http.StatusInternalServerError)
		return
	}

	response := mappers.ToNewGameResponse(game)
	sendJSON(w, response, http.StatusCreated)
}
func (h *GameHandler) MakeMove(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("id")

	var req webmodels.GameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	playerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		sendError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// err := h.service.MakeMove(
	// 	r.Context(),
	// 	gameID,
	// 	playerID,
	// 	req.Row,
	// 	req.Col,
	// )
	// if err != nil {
	// 	sendError(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	err := h.service.MakeMove(r.Context(), gameID, playerID, req.Row, req.Col)
	if err != nil {
		if errors.Is(err, models.ErrGameNotFound) {
			sendError(w, "game not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, models.ErrGameFinished) ||
			errors.Is(err, models.ErrGameNotStarted) ||
			errors.Is(err, models.ErrPlayerNotFound) ||
			errors.Is(err, models.ErrNotPlayerTurn) ||
			errors.Is(err, models.ErrInvalidPosition) ||
			errors.Is(err, models.ErrCellOccupied) {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}

		sendError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	game, err := h.service.GetGame(r.Context(), gameID)
	if err != nil {
		sendError(w, "failed to get game", http.StatusInternalServerError)
		return
	}

	response := mappers.ToWebResponse(game)
	sendJSON(w, response, http.StatusOK)
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("id")

	game, err := h.service.GetGame(r.Context(), gameID)
	if err != nil {
		if errors.Is(err, models.ErrGameNotFound) {
			sendError(w, "game not found", http.StatusNotFound)
			return
		}

		sendError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	response := mappers.ToWebResponse(game)
	sendJSON(w, response, http.StatusOK)
}

func (h *GameHandler) GetAvailableGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.service.GetAvailableGames(r.Context())
	if err != nil {
		sendError(w, "failed to get available games", http.StatusInternalServerError)
		return
	}

	response := make([]any, 0, len(games))
	for _, game := range games {
		response = append(response, mappers.ToWebResponse(game))
	}

	sendJSON(w, response, http.StatusOK)
}

func (h *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("id")

	playerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		sendError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		Cell int `json:"cell"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cell := models.Cell(req.Cell)
	if cell != models.X && cell != models.O {
		sendError(w, "cell must be 1 (X) or 2 (O)", http.StatusBadRequest)
		return
	}

	if err := h.service.JoinGame(
		r.Context(),
		gameID,
		playerID,
		cell,
	); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGame(r.Context(), gameID)
	if err != nil {
		sendError(w, "failed to get game", http.StatusInternalServerError)
		return
	}

	response := mappers.ToWebResponse(game)
	sendJSON(w, response, http.StatusOK)
}

func (h *GameHandler) GetGameHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		sendError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	games, err := h.service.GetGameHistory(r.Context(), userID)
	if err != nil {
		sendError(w, "failed to get game history", http.StatusInternalServerError)
		return
	}

	response := make([]any, 0, len(games))
	for _, game := range games {
		response = append(response, mappers.ToWebResponse(game))
	}

	sendJSON(w, response, http.StatusOK)
}

func (h *GameHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")

	if limitStr == "" {
		sendError(w, "empty limit", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitStr)

	if err != nil {
		sendError(w, "limit must be number", http.StatusBadRequest)
		return
	}

	if limit <= 0 {
		sendError(w, "limit must be positive", http.StatusBadRequest)
		return
	}

	if limit > 100 {
		limit = 100
	}

	items, err := h.service.GetLeaderboard(r.Context(), limit)

	if err != nil {
		sendError(w, "failed to get leaderboard", http.StatusInternalServerError)
		return
	}

	response := make([]webmodels.LeaderboardEntryResponse, 0, len(items))

	for _, entry := range items {
		response = append(response, mappers.ToLeaderboardResponse(entry))
	}

	sendJSON(w, response, http.StatusOK)

}

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

func sendError(w http.ResponseWriter, message string, status int) {
	sendJSON(w, map[string]string{
		"error": message,
	}, status)
}
