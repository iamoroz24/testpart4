package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"project03/internal/domain/models"
	"project03/internal/domain/service"
	"project03/internal/web/middleware"
	"project03/internal/web/webmodels"
)

type AuthHandler struct {
	userService service.UserService
	authService service.AuthService
}

func NewAuthHandler(userService service.UserService, authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
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
	var req webmodels.JwtRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			sendError(w, "invalid login or password", http.StatusUnauthorized)
			return
		}

		sendError(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	response := webmodels.JwtResponse{
		Type:         "Bearer",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	sendJSON(w, response, http.StatusOK)
}

// func (h *AuthHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
// 	var req webmodels.RefreshJwtRequest

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		sendError(w, "invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	if req.RefreshToken == "" {
// 		sendError(w, "empty refresh token", http.StatusBadRequest)
// 		return
// 	}

// 	tokens, err := h.authService.RefreshAccessToken(r.Context(), req.RefreshToken)

// 	if err != nil {
// 		if errors.Is(err, service.ErrInvalidToken) {
// 			sendError(w, "invalid token", http.StatusUnauthorized)
// 			return
// 		}

// 		sendError(w, "internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	response := webmodels.JwtResponse{
// 		Type:         "Bearer",
// 		AccessToken:  tokens.AccessToken,
// 		RefreshToken: tokens.RefreshToken,
// 	}

// 	sendJSON(w, response, http.StatusOK)

// }

func (h *AuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request,
	fn func(context.Context, string) (service.TokenPair, error)) {

	var req webmodels.RefreshJwtRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		sendError(w, "empty refresh token", http.StatusBadRequest)
		return
	}

	tokens, err := fn(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			sendError(w, "invalid token", http.StatusUnauthorized)
			return
		}
		sendError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := webmodels.JwtResponse{
		Type:         "Bearer",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	sendJSON(w, response, http.StatusOK)
}

func (h *AuthHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	h.handleRefreshToken(w, r, h.authService.RefreshAccessToken)
}

func (h *AuthHandler) RefreshRefreshToken(w http.ResponseWriter, r *http.Request) {
	h.handleRefreshToken(w, r, h.authService.RefreshRefreshToken)
}


func(h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request){

	id, ok := middleware.UserIDFromContext(r.Context())

	if !ok {
		sendError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetByID(r.Context(), id.String())

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			sendError(w, "user not found", http.StatusNotFound)
			return
		}

		sendError(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	response := webmodels.UserResponse{
		ID: id,
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
