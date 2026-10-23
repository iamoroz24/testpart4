package di

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"project03/internal/datasource"
	"project03/internal/datasource/repository"
	"project03/internal/domain/service"
	"project03/internal/domain/strategy"
	"project03/internal/web/handlers"
	"project03/internal/web/middleware"
)

func BuildContainer() *fx.App {
	return fx.New(
		fx.Provide(
			NewPostgresPool,
			repository.NewGameRepository,
			strategy.NewMinimaxStrategy,
			service.NewService,
			handlers.NewGameHandler,
			http.NewServeMux,
			repository.NewUserRepository,
			service.NewUserService,
			handlers.NewAuthHandler,
		),
		fx.Invoke(registerRoutes),
		fx.Invoke(startServer),
	)
}

func NewPostgresPool(lc fx.Lifecycle) (*pgxpool.Pool, error) {
	pool, err := datasource.NewPostgresPool()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

// func registerRoutes(
// 	mux *http.ServeMux,
// 	gameHandler *handlers.GameHandler,
// 	authHandler *handlers.AuthHandler,
// ) {
// 	mux.HandleFunc("POST /auth/register", authHandler.Register)

// 	mux.HandleFunc("POST /game", gameHandler.CreateGame)
// 	mux.HandleFunc("POST /game/{id}", gameHandler.MakeMove)
// 	mux.HandleFunc("GET /game/{id}", gameHandler.GetGame)
// 	mux.HandleFunc("POST /auth/login", authHandler.Login)
// }

func registerRoutes(
	mux *http.ServeMux,
	gameHandler *handlers.GameHandler,
	authHandler *handlers.AuthHandler,
	userService service.UserService,
) {
	authMiddleware := middleware.NewUserAuthenticator(userService)

	// Публичные маршруты
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Защищённые игровые маршруты
	mux.Handle("POST /game", authMiddleware(
		http.HandlerFunc(gameHandler.CreateGame),
	))
	mux.Handle("POST /game/{id}", authMiddleware(
		http.HandlerFunc(gameHandler.MakeMove),
	))
	mux.Handle("GET /game/{id}", authMiddleware(
		http.HandlerFunc(gameHandler.GetGame),
	))
	mux.Handle("GET /games", authMiddleware(http.HandlerFunc(gameHandler.GetAvailableGames)))
}

func startServer(lc fx.Lifecycle, mux *http.ServeMux) {
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Server starting on :8080")

			go func() {
				if err := server.ListenAndServe(); err != nil &&
					err != http.ErrServerClosed {
					log.Printf("Server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			return server.Shutdown(shutdownCtx)
		},
	})
}
