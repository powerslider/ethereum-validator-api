package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/powerslider/ethereum-validator-api/pkg/blockreward"
	"github.com/powerslider/ethereum-validator-api/pkg/handlers"
	"github.com/stretchr/testify/require"
)

type mockBlockRewardService struct {
	returnError bool
}

func (m *mockBlockRewardService) GetBlockReward(ctx context.Context, slot uint64) (*blockreward.Result, error) {
	if m.returnError {
		return nil, errors.New("block reward service error")
	}

	return &blockreward.Result{
		Status: "vanilla",
		Reward: "1000",
	}, nil
}

type mockSyncDutyService struct {
	returnError bool
}

func (m *mockSyncDutyService) GetSyncDuties(ctx context.Context, slot uint64) ([]string, error) {
	if m.returnError {
		return nil, errors.New("sync duty service error")
	}

	return []string{"0xabc", "0xdef"}, nil
}

func TestHandlers(t *testing.T) {
	t.Parallel()

	type routeSetup struct {
		handler http.HandlerFunc
		path    string
	}

	type testCase struct {
		route      routeSetup
		name       string
		url        string
		expectBody string
		expected   int
	}

	testCases := []testCase{
		// BlockReward tests
		{
			name: "BlockReward BadRequest",
			route: routeSetup{
				path:    "/blockreward/{slot}",
				handler: handlers.GetBlockRewardHandler(&mockBlockRewardService{}),
			},
			url:        "/blockreward/not-a-number",
			expected:   http.StatusBadRequest,
			expectBody: "Invalid slot number",
		},
		{
			name: "BlockReward Success",
			route: routeSetup{
				path:    "/blockreward/{slot}",
				handler: handlers.GetBlockRewardHandler(&mockBlockRewardService{}),
			},
			url:        "/blockreward/123456",
			expected:   http.StatusOK,
			expectBody: "vanilla",
		},
		{
			name: "BlockReward InternalServerError",
			route: routeSetup{
				path:    "/blockreward/{slot}",
				handler: handlers.GetBlockRewardHandler(&mockBlockRewardService{returnError: true}),
			},
			url:        "/blockreward/123456",
			expected:   http.StatusInternalServerError,
			expectBody: "Failed to retrieve block reward",
		},
		// SyncDuties tests
		{
			name: "SyncDuties BadRequest",
			route: routeSetup{
				path:    "/syncduties/{slot}",
				handler: handlers.GetSyncDutiesHandler(&mockSyncDutyService{}),
			},
			url:        "/syncduties/not-a-number",
			expected:   http.StatusBadRequest,
			expectBody: "Invalid slot number",
		},
		{
			name: "SyncDuties Success",
			route: routeSetup{
				path:    "/syncduties/{slot}",
				handler: handlers.GetSyncDutiesHandler(&mockSyncDutyService{}),
			},
			url:        "/syncduties/123456",
			expected:   http.StatusOK,
			expectBody: "0xabc",
		},
		{
			name: "SyncDuties InternalServerError",
			route: routeSetup{
				path:    "/syncduties/{slot}",
				handler: handlers.GetSyncDutiesHandler(&mockSyncDutyService{returnError: true}),
			},
			url:        "/syncduties/123456",
			expected:   http.StatusInternalServerError,
			expectBody: "Failed to retrieve sync duties",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("Testing URL: %s (expecting %d)", tt.url, tt.expected)

			r := mux.NewRouter()
			r.HandleFunc(tt.route.path, tt.route.handler)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			require.Equal(t, tt.expected, resp.Code)
			require.Contains(t, resp.Body.String(), tt.expectBody)
		})
	}
}
