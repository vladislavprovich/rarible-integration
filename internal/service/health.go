package service

import (
	"context"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
	"net/http"
)

func (s *Service) Health(
	_ context.Context,
	_ rarible.HealthRequest,
) (*rarible.HealthResponse, error) {
	return &rarible.HealthResponse{
		Status: http.StatusOK,
	}, nil
}
