package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/unrolled/render"

	"github.com/vladislavprovich/rarible-integration/internal/service"
)

type Handler interface {
	OwnershipByID(w http.ResponseWriter, r *http.Request)
	QueryTraitsWithRarity(w http.ResponseWriter, r *http.Request)
	Health(writer http.ResponseWriter, reader *http.Request)
}

type ServiceHandler struct {
	service service.Service
	logger  *slog.Logger
	cfg     *Config
	render  *render.Render
}

func NewServiceHandler(srv service.Service, logger *slog.Logger, cfg *Config, render *render.Render) *ServiceHandler {
	return &ServiceHandler{
		service: srv,
		logger:  logger,
		cfg:     cfg,
		render:  render,
	}
}

func (h *ServiceHandler) sendJSON(ctx context.Context, w io.Writer, status int, body any) {
	if err := h.render.JSON(w, status, body); err != nil {
		h.logger.ErrorContext(ctx, "render JSON error", slog.Any("error", err))
	}
}
