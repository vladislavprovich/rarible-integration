package service

import "github.com/vladislavprovich/rarible-integration/pkg/client/rarible"

type ConvectorFromClient struct{}

func NewConvectorFromClient() *ConvectorFromClient {
	return &ConvectorFromClient{}
}

func (c *ConvectorFromClient) ConvertFromGetOwnershipByIDRequest(
	resp *rarible.GetNFTOwnershipByIDResponse,
) *GetNFTOwnershipByIDResponse {
	return &GetNFTOwnershipByIDResponse{
		ID:            resp.ID,
		Blockchain:    resp.Blockchain,
		ItemID:        resp.ItemID,
		Contract:      resp.Contract,
		Collection:    resp.Collection,
		TokenID:       resp.TokenID,
		Owner:         resp.Owner,
		Value:         resp.Value,
		CreatedAt:     resp.CreatedAt,
		LastUpdatedAt: resp.LastUpdatedAt,
		Creators:      toServiceCreators(resp.Creators),
		LazyValue:     resp.LazyValue,
		Pending:       resp.Pending,
		OriginOrders:  resp.OriginOrders,
		Version:       resp.Version,
	}
}

func (c *ConvectorFromClient) ConvertFromPostTraitRaritiesRequest(
	req *rarible.QueryTraitsWithRarityResponse,
) *QueryTraitsWithRarityResponse {
	return &QueryTraitsWithRarityResponse{
		Traits: toServiceTrait(req.Traits),
	}
}

// Support func`s toServiceCreators & toServiceTrait.
func toServiceCreators(creators []rarible.Creator) []Creator {
	result := make([]Creator, len(creators))
	for i, c := range creators {
		result[i] = Creator{
			Account: c.Account,
			Value:   c.Value,
		}
	}
	return result
}

func toServiceTrait(trait []rarible.Trait) []Trait {
	result := make([]Trait, len(trait))
	for i, t := range trait {
		result[i] = Trait{
			Key:    t.Key,
			Value:  t.Value,
			Rarity: t.Rarity,
		}
	}
	return result
}
