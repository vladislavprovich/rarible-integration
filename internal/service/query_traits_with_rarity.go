package service

import (
	"context"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
	"log/slog"
)

func (s *Service) QueryTraitsWithRarity(
	ctx context.Context,
	req *rarible.QueryTraitsWithRarityRequest,
) (*rarible.QueryTraitsWithRarityResponse, error) {
	s.logger.InfoContext(ctx, "QueryTraitsWithRarity", slog.Any("req", req))

	clientResp, err := s.client.QueryTraitsWithRarity(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "service client.QueryTraitsWithRarity", slog.Any("error", err))
		return nil, err
	}

	return clientResp, nil
}
