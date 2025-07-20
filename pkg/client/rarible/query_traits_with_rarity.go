package rarible

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (c *BasicClient) QueryTraitsWithRarity(
	ctx context.Context,
	req *QueryTraitsWithRarityRequest,
) (*QueryTraitsWithRarityResponse, error) {
	payloadConfiguration := PayloadParams{
		CollectionID: req.Collection,
		Properties: []PropertyParams{
			{
				Key:   req.Key,
				Value: req.Value,
			},
		},
	}

	jsonPayload, err := json.Marshal(payloadConfiguration)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %w", err)
	}
	payload := bytes.NewReader(jsonPayload)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.URLForGetTraitRarities, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for GetTraitRarities: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", c.cfg.APISecret)

	res, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error doing request for GetTraitRarities: %w", err)
	}

	defer func() {
		if err = res.Body.Close(); err != nil {
			c.logger.ErrorContext(ctx,
				"error closing response body for GetTraitRarities",
				slog.Any("error", err),
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code for GetTraitRarities: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body for GetTraitRarities: %w", err)
	}

	var resp *QueryTraitsWithRarityResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body for GetTraitRarities: %w", err)
	}

	return resp, nil
}
