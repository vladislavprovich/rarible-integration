package handler

import (
	"encoding/json"
	"fmt"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
	"net/http"
)

func (h *ServiceHandler) Health(writer http.ResponseWriter, reader *http.Request) {
	ctx := reader.Context()

	var req rarible.HealthRequest
	if err := json.NewDecoder(reader.Body).Decode(&req); err != nil {
		h.sendJSON(ctx, writer, http.StatusBadRequest, fmt.Errorf("invalid request %s", req))
		return
	}

	resp, err := h.service.Health(ctx, req)
	if err != nil {
		h.sendJSON(ctx, writer, http.StatusInternalServerError, err)
		return
	}

	h.sendJSON(ctx, writer, http.StatusOK, resp)
}
