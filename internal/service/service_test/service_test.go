package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vladislavprovich/rarible-integration/internal/service"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
)

type mockRaribleClient struct {
	mock.Mock
}

func (m *mockRaribleClient) GetOwnershipByID(
	ctx context.Context,
	req *rarible.GetNFTOwnershipByIDRequest,
) (*rarible.GetNFTOwnershipByIDResponse, error) {
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*rarible.GetNFTOwnershipByIDResponse)
	return resp, args.Error(1)
}

func (m *mockRaribleClient) QueryTraitsWithRarity(
	ctx context.Context,
	req *rarible.QueryTraitsWithRarityRequest,
) (*rarible.QueryTraitsWithRarityResponse, error) {
	args := m.Called(ctx, req)
	resp, _ := args.Get(0).(*rarible.QueryTraitsWithRarityResponse)
	return resp, args.Error(1)
}

func TestService_GetOwnershipByID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	now := time.Now()

	tests := []struct {
		name       string
		req        *service.GetNFTOwnershipByIDRequest
		clientResp *rarible.GetNFTOwnershipByIDResponse
		clientErr  error
		wantResp   *service.GetNFTOwnershipByIDResponse
		wantErr    bool
	}{
		{
			name: "success",
			req: &service.GetNFTOwnershipByIDRequest{
				OwnershipID: "ownership123",
			},
			clientResp: &rarible.GetNFTOwnershipByIDResponse{
				ID:            "ownership123",
				Blockchain:    "ethereum",
				ItemID:        "item456",
				Contract:      "0xabc123",
				Collection:    "cool-collection",
				TokenID:       "token789",
				Owner:         "0xowneraddress",
				Value:         "1",
				CreatedAt:     now,
				LastUpdatedAt: now,
				Creators: []rarible.Creator{
					{Account: "0xcreator1", Value: 50},
					{Account: "0xcreator2", Value: 50},
				},
				LazyValue:    "0",
				Pending:      []any{},
				OriginOrders: []any{},
				Version:      1,
			},
			wantResp: &service.GetNFTOwnershipByIDResponse{
				ID:            "ownership123",
				Blockchain:    "ethereum",
				ItemID:        "item456",
				Contract:      "0xabc123",
				Collection:    "cool-collection",
				TokenID:       "token789",
				Owner:         "0xowneraddress",
				Value:         "1",
				CreatedAt:     now,
				LastUpdatedAt: now,
				Creators: []service.Creator{
					{Account: "0xcreator1", Value: 50},
					{Account: "0xcreator2", Value: 50},
				},
				LazyValue:    "0",
				Pending:      []any{},
				OriginOrders: []any{},
				Version:      1,
			},
			wantErr: false,
		},
		{
			name:       "client error",
			req:        &service.GetNFTOwnershipByIDRequest{OwnershipID: "ownership999"},
			clientResp: nil,
			clientErr:  errors.New("client failure"),
			wantResp:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockRaribleClient)
			mockClient.On("GetOwnershipByID", mock.Anything, mock.Anything).
				Return(tt.clientResp, tt.clientErr)

			svc := service.NewRaribleService(context.Background(), logger, mockClient)

			gotResp, err := svc.GetOwnershipByID(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.wantResp, gotResp)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestService_PostTraitRarities(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name       string
		req        *service.QueryTraitsWithRarityRequest
		clientResp *rarible.QueryTraitsWithRarityResponse
		clientErr  error
		wantResp   *service.QueryTraitsWithRarityResponse
		wantErr    bool
	}{
		{
			name: "success",
			req: &service.QueryTraitsWithRarityRequest{
				Collection: "cool-collection",
				Key:        "color",
				Value:      "red",
			},
			clientResp: &rarible.QueryTraitsWithRarityResponse{
				Traits: []rarible.Trait{
					{Key: "color", Value: "red", Rarity: "rare"},
					{Key: "size", Value: "large", Rarity: "common"},
				},
			},
			wantResp: &service.QueryTraitsWithRarityResponse{
				Traits: []service.Trait{
					{Key: "color", Value: "red", Rarity: "rare"},
					{Key: "size", Value: "large", Rarity: "common"},
				},
			},
			wantErr: false,
		},
		{
			name: "client error",
			req: &service.QueryTraitsWithRarityRequest{
				Collection: "cool-collection",
				Key:        "color",
				Value:      "blue",
			},
			clientResp: nil,
			clientErr:  errors.New("client failure"),
			wantResp:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockRaribleClient)
			mockClient.On("QueryTraitsWithRarity", mock.Anything, mock.Anything).
				Return(tt.clientResp, tt.clientErr)

			svc := service.NewRaribleService(context.Background(), logger, mockClient)

			gotResp, err := svc.QueryTraitsWithRarity(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				assert.Equal(t, tt.wantResp, gotResp)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
