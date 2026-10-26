package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"github.com/google/uuid"

	"project03/internal/domain/models"
	"project03/internal/domain/service"
	"project03/internal/web/webmodels"
)

type AuthHandler struct {
	userService service.UserService
}

func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req webmodels.SignUpRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(
		r.Context(),
		req.Login,
		req.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidLogin),
			errors.Is(err, service.ErrInvalidPassword):
			sendError(w, err.Error(), http.StatusBadRequest)

		case errors.Is(err, models.ErrLoginAlreadyTaken):
			sendError(w, "login already taken", http.StatusConflict)

		default:
			sendError(w, "failed to register user", http.StatusInternalServerError)
		}
		return
	}

	response := webmodels.UserResponse{
		ID:    user.ID,
		Login: user.Login,
	}

	sendJSON(w, response, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	login, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="tictactoe"`)
		sendError(w, "authentication required", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), login, password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.Header().Set("WWW-Authenticate", `Basic realm="tictactoe"`)
			sendError(w, "invalid login or password", http.StatusUnauthorized)
			return
		}

		sendError(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	response := webmodels.UserResponse{
		ID:    user.ID,
		Login: user.Login,
	}
	sendJSON(w, response, http.StatusOK)
}

func (h *AuthHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, err := uuid.Parse(id); err != nil {
		sendError(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			sendError(w, "user not found", http.StatusNotFound)
			return
		}

		sendError(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	response := webmodels.UserResponse{
		ID:    user.ID,
		Login: user.Login,
	}

	sendJSON(w, response, http.StatusOK)
}
