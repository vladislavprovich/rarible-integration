package service

import (
	"context"
	"net/http"

	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
)

func (s *Service) Health(
	_ context.Context,
	_ rarible.HealthRequest,
) (*rarible.HealthResponse, error) {
	return &rarible.HealthResponse{
		Status: http.StatusOK,
	}, nil
}
