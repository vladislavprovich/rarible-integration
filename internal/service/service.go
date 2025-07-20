package service

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
)

type RaribleService interface {
	GetOwnershipByID(
		ctx context.Context,
		req *GetNFTOwnershipByIDRequest,
	) (*GetNFTOwnershipByIDResponse, error)
	QueryTraitsWithRarity(
		ctx context.Context,
		req *QueryTraitsWithRarityRequest,
	) (*QueryTraitsWithRarityResponse, error)
}

type Service struct {
	logger              *slog.Logger
	client              rarible.Client
	convectorToClient   *ConvectorToClient
	convectorFromClient *ConvectorFromClient
}

func NewRaribleService(_ context.Context, log *slog.Logger, client rarible.Client) *Service {
	return &Service{
		logger:              log,
		client:              client,
		convectorToClient:   NewConvectorToClient(),
		convectorFromClient: NewConvectorFromClient(),
	}
}

func (s *Service) Health(
	_ context.Context,
	_ rarible.HealthRequest,
) (*rarible.HealthResponse, error) {
	return &rarible.HealthResponse{
		Status: http.StatusOK,
	}, nil
}
