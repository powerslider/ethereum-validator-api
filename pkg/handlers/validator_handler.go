package handlers

import (
	"context"
	"encoding/json"
	"errors"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
	"github.com/powerslider/ethereum-validator-api/pkg/blockreward"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// BlockRewardService defines a minimal interface for block reward operations.
type BlockRewardService interface {
	GetBlockReward(ctx context.Context, slot uint64) (*blockreward.Result, error)
}

// SyncDutyService defines a minimal interface for sync duties operations.
type SyncDutyService interface {
	GetSyncDuties(ctx context.Context, slot uint64) ([]string, error)
}

// blockRewardResponse defines the structure returned for block reward lookup.
type blockRewardResponse struct {
	Status string `json:"status"`
	Reward string `json:"reward"`
}

// syncDutiesResponse defines the structure returned for sync duties lookup.
type syncDutiesResponse struct {
	Validators []string `json:"validators"`
}

// GetBlockRewardHandler handles block reward lookup.
// @Summary Get Block Reward
// @Description Retrieves block reward details for a given slot.
// @Tags BlockReward
// @Accept json
// @Produce json
// @Param slot path int true "Slot number"
// @Success 200 {object} blockRewardResponse
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /blockreward/{slot} [get]
func GetBlockRewardHandler(svc BlockRewardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slotStr := mux.Vars(r)["slot"]

		slot, err := strconv.ParseUint(slotStr, 10, 64)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "Invalid slot number", err)
			return
		}

		result, err := svc.GetBlockReward(r.Context(), slot)
		if err != nil {
			wrappedErr := err

			switch e := pkgerrors.Cause(wrappedErr); {
			case errors.Is(e, beacon.ErrSlotMissedOrDoesNotExist):
				writeAPIError(w, http.StatusNotFound, "Slot was missed", wrappedErr)
			case errors.Is(e, beacon.ErrSlotInFuture):
				writeAPIError(w, http.StatusBadRequest, "Slot is in the future", wrappedErr)
			default:
				writeAPIError(w, http.StatusInternalServerError, "Failed to retrieve block reward", wrappedErr)
			}

			return
		}

		resp := blockRewardResponse{
			Status: result.Status,
			Reward: result.Reward,
		}

		w.Header().Set("Content-Type", "application/json")

		if err = json.NewEncoder(w).Encode(resp); err != nil {
			return
		}
	}
}

// GetSyncDutiesHandler handles sync committee duties lookup.
// @Summary Get Sync Duties
// @Description Retrieves validators assigned for sync committee duties for a given slot.
// @Tags SyncDuties
// @Accept json
// @Produce json
// @Param slot path int true "Slot number"
// @Success 200 {object} syncDutiesResponse
// @Failure 400 {object} APIError
// @Failure 500 {object} APIError
// @Router /syncduties/{slot} [get]
func GetSyncDutiesHandler(svc SyncDutyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slotStr := mux.Vars(r)["slot"]

		slot, err := strconv.ParseUint(slotStr, 10, 64)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "Invalid slot number", err)
			return
		}

		validators, err := svc.GetSyncDuties(r.Context(), slot)
		if err != nil {
			wrappedErr := err

			switch e := pkgerrors.Cause(wrappedErr); {
			case errors.Is(e, beacon.ErrDutiesNotFound):
				writeAPIError(w, http.StatusNotFound, "Sync duties not found", wrappedErr)
			case errors.Is(e, beacon.ErrSlotInFuture):
				writeAPIError(w, http.StatusBadRequest, "Slot is in the future", wrappedErr)
			case errors.Is(e, beacon.ErrSlotWasMissed):
				writeAPIError(w, http.StatusBadRequest, "Slot was missed", wrappedErr)
			default:
				writeAPIError(w, http.StatusInternalServerError, "Failed to retrieve sync duties", wrappedErr)
			}

			return
		}

		resp := syncDutiesResponse{
			Validators: validators,
		}

		w.Header().Set("Content-Type", "application/json")

		if err = json.NewEncoder(w).Encode(resp); err != nil {
			return
		}
	}
}
