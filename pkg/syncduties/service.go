package syncduties

import (
	"context"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
)

type BeaconService interface {
	GetCurrentSlot(ctx context.Context) (uint64, error)
	FetchSyncCommitteeIndexes(ctx context.Context, slot uint64) ([]string, error)
	FetchValidatorsByIDs(ctx context.Context, slot uint64, ids []string) ([]string, error)
}

type Service struct {
	BeaconService BeaconService
}

func NewService(svc BeaconService) *Service {
	return &Service{
		BeaconService: svc,
	}
}

func (s *Service) GetSyncDuties(ctx context.Context, slot uint64) ([]string, error) {
	// Step 0: Validate if slot is in the future.
	currentSlot, err := s.BeaconService.GetCurrentSlot(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch current slot")
	}

	if slot > currentSlot+1 {
		return nil, beacon.ErrSlotInFuture
	}

	validatorIndexes, err := s.BeaconService.FetchSyncCommitteeIndexes(ctx, slot)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch sync committee")
	}

	validators, err := s.BeaconService.FetchValidatorsByIDs(ctx, slot, validatorIndexes)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch validators by ids")
	}

	return validators, nil
}
