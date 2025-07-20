package service

import (
	"context"
	"log/slog"
)

func (s *Service) GetOwnershipByID(
	ctx context.Context,
	req *GetNFTOwnershipByIDRequest,
) (*GetNFTOwnershipByIDResponse, error) {
	s.logger.InfoContext(ctx, "GetOwnershipByID", slog.Any("req", req))
	integrationReq := s.convectorToClient.ConvertToGetOwnershipByIDRequest(req)

	clientResp, err := s.client.GetOwnershipByID(ctx, integrationReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "service client.GetOwnershipByID", slog.Any("error", err))
		return nil, err
	}

	resp := s.convectorFromClient.ConvertFromGetOwnershipByIDRequest(clientResp)

	return resp, nil
}
