package client_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
)

func TestBasicClient_GetOwnershipByID(t *testing.T) {
	tests := []struct {
		name               string
		ownershipID        string
		mockStatusCode     int
		mockResponseBody   string
		expectedErr        string
		expectedParsedResp *rarible.GetNFTOwnershipByIDResponse
	}{
		{
			name:           "valid_response",
			ownershipID:    "valid-id",
			mockStatusCode: http.StatusOK,
			mockResponseBody: fmt.Sprintf(`{
				"id": "valid-id",
				"blockchain": "ETHEREUM",
				"itemId": "item-123",
				"contract": "0xabc",
				"collection": "cool-collection",
				"tokenId": "token-999",
				"owner": "0xowner",
				"value": "100",
				"createdAt": "%s",
				"lastUpdatedAt": "%s",
				"creators": [{"account": "0xcreator", "value": 10}],
				"lazyValue": "50",
				"pending": [],
				"originOrders": [],
				"version": 2
			}`, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339)),
			expectedParsedResp: &rarible.GetNFTOwnershipByIDResponse{
				ID:           "valid-id",
				Blockchain:   "ETHEREUM",
				ItemID:       "item-123",
				Contract:     "0xabc",
				Collection:   "cool-collection",
				TokenID:      "token-999",
				Owner:        "0xowner",
				Value:        "100",
				Creators:     []rarible.Creator{{Account: "0xcreator", Value: 10}},
				LazyValue:    "50",
				Pending:      []any{},
				OriginOrders: []any{},
				Version:      2,
			},
		},
		{
			name:             "not_found_with_api_error",
			ownershipID:      "not-found-id",
			mockStatusCode:   http.StatusNotFound,
			mockResponseBody: `{"code":"NOT_FOUND","message":"Ownership not found"}`,
			expectedErr:      "Ownership not found",
		},
		{
			name:             "invalid_json_response",
			ownershipID:      "bad-json",
			mockStatusCode:   http.StatusOK,
			mockResponseBody: `{invalid-json}`,
			expectedErr:      "error unmarshalling response body for GetOwnershipByID",
		},
		{
			name:             "error_parsing_error_body",
			ownershipID:      "bad-error-body",
			mockStatusCode:   http.StatusBadRequest,
			mockResponseBody: `non-json-error`,
			expectedErr:      "unexpected status 400 and cannot parse error body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGetOwnershipByIDTest(t, tt)
		})
	}
}

func runGetOwnershipByIDTest(t *testing.T, tt struct {
	name               string
	ownershipID        string
	mockStatusCode     int
	mockResponseBody   string
	expectedErr        string
	expectedParsedResp *rarible.GetNFTOwnershipByIDResponse
}) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(tt.mockStatusCode)
		_, _ = w.Write([]byte(tt.mockResponseBody))
	}))
	defer server.Close()

	cfg := &rarible.Config{
		APISecret:                 "test-api-key",
		URLForGetNFTOwnershipByID: server.URL,
		URLForGetTraitRarities:    server.URL,
	}

	client := rarible.NewBasicClient(server.Client(), cfg, slog.Default())

	req := &rarible.GetNFTOwnershipByIDRequest{OwnershipID: tt.ownershipID}
	resp, err := client.GetOwnershipByID(context.Background(), req)

	if tt.expectedErr != "" {
		if err == nil {
			t.Fatalf("expected error %q but got nil", tt.expectedErr)
		}
		if !strings.Contains(err.Error(), tt.expectedErr) {
			t.Errorf("expected error to contain %q, got %q", tt.expectedErr, err.Error())
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if !equalOwnershipResponses(resp, tt.expectedParsedResp) {
		t.Errorf("unexpected response:\ngot  %+v\nwant %+v", resp, tt.expectedParsedResp)
	}
}

func TestBasicClient_QueryTraitsWithRarity(t *testing.T) {
	tests := []struct {
		name             string
		mockStatusCode   int
		mockResponseBody string
		request          *rarible.QueryTraitsWithRarityRequest
		expectedErr      string
		expectedResponse *rarible.QueryTraitsWithRarityResponse
	}{
		{
			name:           "valid_response",
			mockStatusCode: http.StatusOK,
			mockResponseBody: `{
				"traits": [
					{
						"key": "color",
						"value": "red",
						"rarity": "0.05"
					},
					{
						"key": "size",
						"value": "large",
						"rarity": "0.2"
					}
				]
			}`,
			request: &rarible.QueryTraitsWithRarityRequest{
				Collection: "col-1",
				Key:        "color",
				Value:      "red",
			},
			expectedResponse: &rarible.QueryTraitsWithRarityResponse{
				Traits: []rarible.Trait{
					{Key: "color", Value: "red", Rarity: "0.05"},
					{Key: "size", Value: "large", Rarity: "0.2"},
				},
			},
		},
		{
			name:             "non_200_status",
			mockStatusCode:   http.StatusBadRequest,
			mockResponseBody: `bad request`,
			request: &rarible.QueryTraitsWithRarityRequest{
				Collection: "col-2",
				Key:        "shape",
				Value:      "circle",
			},
			expectedErr: "unexpected status code for GetTraitRarities",
		},
		{
			name:             "invalid_json_response",
			mockStatusCode:   http.StatusOK,
			mockResponseBody: `{invalid-json}`,
			request: &rarible.QueryTraitsWithRarityRequest{
				Collection: "col-3",
				Key:        "material",
				Value:      "metal",
			},
			expectedErr: "error unmarshalling response body for GetTraitRarities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runQueryTraitsWithRarityTest(t, tt)
		})
	}
}

func runQueryTraitsWithRarityTest(t *testing.T, tt struct {
	name             string
	mockStatusCode   int
	mockResponseBody string
	request          *rarible.QueryTraitsWithRarityRequest
	expectedErr      string
	expectedResponse *rarible.QueryTraitsWithRarityResponse
}) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method but got %s", r.Method)
		}
		w.WriteHeader(tt.mockStatusCode)
		_, _ = w.Write([]byte(tt.mockResponseBody))
	}))
	defer server.Close()

	cfg := &rarible.Config{
		APISecret:                 "test-api-key",
		URLForGetNFTOwnershipByID: server.URL,
		URLForGetTraitRarities:    server.URL,
	}

	client := rarible.NewBasicClient(server.Client(), cfg, slog.Default())

	resp, err := client.QueryTraitsWithRarity(context.Background(), tt.request)

	if tt.expectedErr != "" {
		if err == nil {
			t.Fatalf("expected error %q but got nil", tt.expectedErr)
		}
		if !strings.Contains(err.Error(), tt.expectedErr) {
			t.Errorf("expected error to contain %q, got %q", tt.expectedErr, err.Error())
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if !equalTraitsWithRarityResponse(resp, tt.expectedResponse) {
		t.Errorf("unexpected response:\ngot  %+v\nwant %+v", resp, tt.expectedResponse)
	}
}

// Creator comparison.
func equalCreators(a, b []rarible.Creator) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Comparison of the main string/int fields GetNFTOwnershipByIDResponse.
func equalBasicOwnershipFields(a, b *rarible.GetNFTOwnershipByIDResponse) bool {
	return a.ID == b.ID &&
		a.Blockchain == b.Blockchain &&
		a.ItemID == b.ItemID &&
		a.Contract == b.Contract &&
		a.Collection == b.Collection &&
		a.TokenID == b.TokenID &&
		a.Owner == b.Owner &&
		a.Value == b.Value &&
		a.LazyValue == b.LazyValue &&
		a.Version == b.Version
}

// Helper for comparison GetNFTOwnershipByIDResponse (without time.Time fields).
func equalOwnershipResponses(a, b *rarible.GetNFTOwnershipByIDResponse) bool {
	if a == nil || b == nil {
		return a == b
	}
	if !equalBasicOwnershipFields(a, b) {
		return false
	}
	if !equalCreators(a.Creators, b.Creators) {
		return false
	}
	return true
}

// Trait comparison.
func equalTraits(a, b []rarible.Trait) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Key != b[i].Key ||
			a[i].Value != b[i].Value ||
			a[i].Rarity != b[i].Rarity {
			return false
		}
	}
	return true
}

// Helper for comparing QueryTraitsWithRarityResponse.
func equalTraitsWithRarityResponse(a, b *rarible.QueryTraitsWithRarityResponse) bool {
	if a == nil || b == nil {
		return a == b
	}
	return equalTraits(a.Traits, b.Traits)
}
