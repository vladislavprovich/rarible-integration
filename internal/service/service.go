package service

import (
	"context"
	"log/slog"

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
