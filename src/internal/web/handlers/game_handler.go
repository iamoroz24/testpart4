package handlers

import (
	"encoding/json"
	"net/http"


	"project03/internal/domain/models"
	"project03/internal/domain/service"
	"project03/internal/web/mappers"
	"project03/internal/web/webmodels"
	"project03/internal/web/middleware"
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
	

	err := h.service.MakeMove(
		r.Context(),
		gameID,
		playerID,
		req.Row,
		req.Col,
	)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	game, err := h.service.GetGame(r.Context(), gameID)
	if err != nil {
		sendError(w, "failed to get game", http.StatusInternalServerError)
		return
	}

	response := mappers.ToWebResponse(game, "")
	sendJSON(w, response, http.StatusOK)
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("id")

	game, err := h.service.GetGame(r.Context(), gameID)
	if err != nil {
		sendError(w, "game not found", http.StatusNotFound)
		return
	}

	response := mappers.ToWebResponse(game, "")
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
		response = append(response, mappers.ToWebResponse(game, ""))
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

	response := mappers.ToWebResponse(game, "")
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
