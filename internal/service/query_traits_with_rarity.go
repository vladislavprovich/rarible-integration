package service

import (
	"context"
	"log/slog"
)

func (s *Service) QueryTraitsWithRarity(
	ctx context.Context,
	req *QueryTraitsWithRarityRequest,
) (*QueryTraitsWithRarityResponse, error) {
	s.logger.InfoContext(ctx, "QueryTraitsWithRarity", slog.Any("req", req))
	integrationReq := s.convectorToClient.ConvertToPostTraitRaritiesRequest(req)

	clientResp, err := s.client.QueryTraitsWithRarity(ctx, integrationReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "service client.QueryTraitsWithRarity", slog.Any("error", err))
		return nil, err
	}

	resp := s.convectorFromClient.ConvertFromPostTraitRaritiesRequest(clientResp)

	return resp, nil
}
