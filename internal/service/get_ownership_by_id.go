package service

import (
	"context"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
	"log/slog"
)

func (s *Service) GetOwnershipByID(
	ctx context.Context,
	req *rarible.GetNFTOwnershipByIDRequest,
) (*rarible.GetNFTOwnershipByIDResponse, error) {
	s.logger.InfoContext(ctx, "GetOwnershipByID", slog.Any("req", req))

	clientResp, err := s.client.GetOwnershipByID(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "service client.GetOwnershipByID", slog.Any("error", err))
		return nil, err
	}

	return clientResp, nil
}
