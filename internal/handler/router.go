package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	logger2 "github.com/vladislavprovich/rarible-integration/pkg/logger"

	"github.com/go-chi/cors"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(handler Handler, logger *slog.Logger, cfg *Config) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.Timeout(cfg.Timeout))

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"Accept", "Content-Type", "authorization"},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           int(cfg.MaxAge),
	}))

	wrappedLogger := &logger2.Logger{Logger: logger}
	mux.Use(chiMiddleware.RequestID)
	mux.Use(chiMiddleware.Logger)
	mux.Use(chiMiddleware.RequestLogger(&chiMiddleware.DefaultLogFormatter{
		Logger:  wrappedLogger,
		NoColor: false,
	}))

	routUrl := fmt.Sprintf("api/%s/rarible", cfg.ApiVersion)
	mux.Route(routUrl, func(r chi.Router) {
		r.Get("/ownership", handler.OwnershipByID)
		r.Post("/rarities", handler.QueryTraitsWithRarity)
	})

	return mux
}
