package rarible

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func (c *BasicClient) GetOwnershipByID(
	ctx context.Context,
	req *GetNFTOwnershipByIDRequest,
) (*GetNFTOwnershipByIDResponse, error) {
	urlForGetOwnershipByID := fmt.Sprintf("%s/%s",
		c.cfg.URLForGetNFTOwnershipByID,
		req.OwnershipID,
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, urlForGetOwnershipByID, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request for GetOwnershipByID: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-Api-Key", c.cfg.APISecret)

	res, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error doing request for GetOwnershipByID: %w", err)
	}

	defer func() {
		if err = res.Body.Close(); err != nil {
			c.logger.ErrorContext(ctx,
				"error closing response body for GetOwnershipByID",
				slog.Any("error", err),
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)

		var apiErr APIError
		if err = json.Unmarshal(body, &apiErr); err != nil {
			// If the body is not parsed, we return a general error.
			return nil, fmt.Errorf("unexpected status %d and cannot parse error body: %s",
				res.StatusCode,
				string(body),
			)
		}

		apiErr.StatusCode = int64(res.StatusCode)
		return nil, &apiErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body for GetOwnershipByID: %w", err)
	}

	var resp *GetNFTOwnershipByIDResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body for GetOwnershipByID: %w", err)
	}

	return resp, nil
}
