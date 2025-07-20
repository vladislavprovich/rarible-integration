package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/vladislavprovich/rarible-integration/internal/service"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
)

func (h *ServiceHandler) OwnershipByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req service.GetNFTOwnershipByIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.ErrorContext(ctx, "failed to decode request body", slog.Any("error", err))
		h.sendJSON(ctx, w, http.StatusBadRequest, err)
		return
	}

	resp, err := h.service.GetOwnershipByID(ctx, &req)
	if err != nil {
		var apiErr *rarible.APIError
		if errors.As(err, &apiErr) {
			h.logger.WarnContext(ctx, "client validation error", slog.Any("error", err))
			h.sendJSON(ctx, w, http.StatusInternalServerError, err)
			return
		}

		h.logger.ErrorContext(ctx, "GetOwnershipByID error", slog.Any("error", err))
		h.sendJSON(ctx, w, http.StatusInternalServerError, err)
		return
	}

	h.sendJSON(ctx, w, http.StatusOK, resp)
}
