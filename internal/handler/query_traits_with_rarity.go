package handler

import (
	"encoding/json"
	"errors"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
	"log/slog"
	"net/http"
)

func (h *ServiceHandler) QueryTraitsWithRarity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req rarible.QueryTraitsWithRarityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.ErrorContext(ctx, "decode QueryTraitsWithRarity error", slog.Any("error", err))
		h.sendJSON(ctx, w, http.StatusBadRequest, err)
		return
	}

	resp, err := h.service.QueryTraitsWithRarity(ctx, &req)
	if err != nil {
		var apiErr *rarible.APIError
		if errors.As(err, &apiErr) {
			h.logger.WarnContext(ctx, "client validation error", slog.Any("error", err))
			h.sendJSON(ctx, w, http.StatusInternalServerError, err)
			return
		}

		h.logger.ErrorContext(ctx, "QueryTraitsWithRarity error", slog.Any("error", err))
		h.sendJSON(ctx, w, http.StatusInternalServerError, err)
		return
	}

	h.sendJSON(ctx, w, http.StatusOK, resp)
}
