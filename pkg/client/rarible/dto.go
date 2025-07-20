package rarible

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	GetNFTOwnershipByIDRequest struct {
		OwnershipID string `json:"ownershipId"`
	}

	GetNFTOwnershipByIDResponse struct {
		ID            string    `json:"id"`
		Blockchain    string    `json:"blockchain"`
		ItemID        string    `json:"itemId"`
		Contract      string    `json:"contract"`
		Collection    string    `json:"collection"`
		TokenID       string    `json:"tokenId"`
		Owner         string    `json:"owner"`
		Value         string    `json:"value"`
		CreatedAt     time.Time `json:"createdAt"`
		LastUpdatedAt time.Time `json:"lastUpdatedAt"`
		Creators      []Creator `json:"creators"`
		LazyValue     string    `json:"lazyValue"`
		Pending       []any     `json:"pending"`
		OriginOrders  []any     `json:"originOrders"`
		Version       int64     `json:"version"`
	}

	Creator struct {
		Account string `json:"account"`
		Value   int64  `json:"value"`
	}
)

type (
	QueryTraitsWithRarityRequest struct {
		Collection string `json:"collection"`
		Key        string `json:"key"`
		Value      string `json:"value"`
	}

	QueryTraitsWithRarityResponse struct {
		Traits []Trait `json:"traits"`
	}

	Trait struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Rarity string `json:"rarity"`
	}
)

type (
	// PropertyParams PayloadParams Used in the PostTraitRarities method, to create JSON payload.
	PropertyParams struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	PayloadParams struct {
		CollectionID string           `json:"collectionId"`
		Properties   []PropertyParams `json:"properties"`
	}
)

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int64  `json:"-"`
}

type (
	HealthRequest struct {
	}

	HealthResponse struct {
		Status int `json:"status"`
	}
)

func (r *GetNFTOwnershipByIDRequest) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, r,
		validation.Field(r.OwnershipID),
	)
}

func (r QueryTraitsWithRarityRequest) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, r,
		validation.Field(r.Value),
		validation.Field(r.Key),
		validation.Field(r.Collection),
	)
}
