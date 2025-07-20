package service

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
		ID            string
		Blockchain    string
		ItemID        string
		Contract      string
		Collection    string
		TokenID       string
		Owner         string
		Value         string
		CreatedAt     time.Time
		LastUpdatedAt time.Time
		Creators      []Creator
		LazyValue     string
		Pending       []any
		OriginOrders  []any
		Version       int64
	}

	Creator struct {
		Account string
		Value   int64
	}
)

type (
	QueryTraitsWithRarityRequest struct {
		Key        string `json:"key"`
		Value      string `json:"value"`
		Collection string `json:"collection"`
	}

	QueryTraitsWithRarityResponse struct {
		Traits []Trait
	}

	Trait struct {
		Key    string
		Value  string
		Rarity string
	}
)

func (r *QueryTraitsWithRarityRequest) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, r,
		validation.Field(r.Value, validation.Required),
		validation.Field(r.Key, validation.Required),
		validation.Field(r.Collection, validation.Required),
	)
}

func (r *GetNFTOwnershipByIDRequest) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, r,
		validation.Field(r.OwnershipID, validation.Required),
	)
}
