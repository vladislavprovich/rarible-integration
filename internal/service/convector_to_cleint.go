package service

import "github.com/vladislavprovich/rarible-integration/pkg/client/rarible"

type ConvectorToClient struct{}

func NewConvectorToClient() *ConvectorToClient {
	return &ConvectorToClient{}
}

func (c *ConvectorToClient) ConvertToGetOwnershipByIDRequest(
	req *GetNFTOwnershipByIDRequest,
) *rarible.GetNFTOwnershipByIDRequest {
	return &rarible.GetNFTOwnershipByIDRequest{
		OwnershipID: req.OwnershipID,
	}
}

func (c *ConvectorToClient) ConvertToPostTraitRaritiesRequest(
	req *QueryTraitsWithRarityRequest,
) *rarible.QueryTraitsWithRarityRequest {
	return &rarible.QueryTraitsWithRarityRequest{
		Collection: req.Collection,
		Key:        req.Key,
		Value:      req.Value,
	}
}
