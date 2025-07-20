package rarible

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	APISecret                 string `envconfig:"RARIBLE_API_SECRET"`
	URLForGetNFTOwnershipByID string `envconfig:"RARIBLE_URL_OWNERSHIP"`
	URLForGetTraitRarities    string `envconfig:"RARIBLE_URL_RARITIES"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.APISecret),
		validation.Field(&c.URLForGetNFTOwnershipByID),
		validation.Field(&c.URLForGetNFTOwnershipByID),
	)
}
