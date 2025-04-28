package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/powerslider/ethereum-validator-api/pkg/blockreward"
	"github.com/powerslider/ethereum-validator-api/pkg/syncduties"
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
			switch {
			case errors.Is(err, blockreward.ErrSlotMissedOrDoesNotExist):
				writeAPIError(w, http.StatusNotFound, "Slot was missed", err)
			case errors.Is(err, blockreward.ErrSlotInFuture):
				writeAPIError(w, http.StatusBadRequest, "Slot is in the future", err)
			default:
				writeAPIError(w, http.StatusInternalServerError, "Failed to retrieve block reward", err)
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
			switch {
			case errors.Is(err, syncduties.ErrDutiesNotFound):
				writeAPIError(w, http.StatusNotFound, "Sync duties not found", err)
			case errors.Is(err, syncduties.ErrSlotTooFarInFuture):
				writeAPIError(w, http.StatusBadRequest, "Slot is too far in the future", err)
			default:
				writeAPIError(w, http.StatusInternalServerError, "Failed to retrieve sync duties", err)
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
