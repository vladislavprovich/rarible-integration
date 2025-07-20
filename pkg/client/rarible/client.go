package rarible

import (
	"context"
	"log/slog"
	"net/http"
)

type Client interface {
	GetOwnershipByID(
		ctx context.Context,
		req *GetNFTOwnershipByIDRequest,
	) (*GetNFTOwnershipByIDResponse, error)
	QueryTraitsWithRarity(
		ctx context.Context,
		req *QueryTraitsWithRarityRequest,
	) (*QueryTraitsWithRarityResponse, error)
}

type BasicClient struct {
	client *http.Client
	logger *slog.Logger
	cfg    *Config
}

func NewBasicClient(httpClient *http.Client, cfg *Config, log *slog.Logger) *BasicClient {
	return &BasicClient{
		client: httpClient,
		logger: log,
		cfg:    cfg,
	}
}
