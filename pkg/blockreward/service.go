package blockreward

import (
	"context"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	statusVanilla = "vanilla"
	statusMEV     = "mev"
)

var mevRelaySignatures = []string{
	"flashbots",
	"bloxroute",
	"eden",
	"manifold",
	"builder0x69",
	"rsync-builder",
	"beaverbuild",
	"aestus",
	"titans",
	"relayooor",
}

type BeaconService interface {
	GetCurrentSlot(ctx context.Context) (uint64, error)
	GetBeaconHeader(ctx context.Context, slot uint64) (*beacon.BlockHeaderResponse, error)
	GetBlockRewardFromConsensus(ctx context.Context, blockRoot string) (*beacon.RewardResponse, error)
}

// Service provides block reward calculation functionality.
type Service struct {
	ExecClient    *ethclient.Client
	BeaconService BeaconService
}

// NewService creates a new block reward service instance.
func NewService(client *ethclient.Client, svc BeaconService) *Service {
	return &Service{
		ExecClient:    client,
		BeaconService: svc,
	}
}

type Result struct {
	Status string
	Reward string
}

// GetBlockReward calculates the block reward earned by the validator at a given slot.
// It returns the block status ("vanilla" or "mev") and the reward amount in Gwei.
func (s *Service) GetBlockReward(ctx context.Context, slot uint64) (*Result, error) {
	// Step 0: Validate if slot is in the future.
	currentSlot, err := s.BeaconService.GetCurrentSlot(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch current slot")
	}

	if slot > currentSlot+1 {
		return nil, beacon.ErrSlotInFuture
	}

	// Step 1: Get block header to retrieve proposer index and block root.
	headerResp, err := s.BeaconService.GetBeaconHeader(ctx, slot)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch block header")
	}

	blockRoot := headerResp.Data.Root

	// Step 2: Get consensus-layer reward (already in Gwei).
	rewardResp, err := s.BeaconService.GetBlockRewardFromConsensus(ctx, blockRoot)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch consensus-layer reward")
	}

	// Step 3: Fetch execution block to inspect ExtraData for MEV tag.
	execBlock, err := s.ExecClient.BlockByNumber(ctx, new(big.Int).SetUint64(slot))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch execution block")
	}

	extra := strings.ToLower(strings.TrimSpace(string(execBlock.Extra())))
	status := statusVanilla

	for _, sig := range mevRelaySignatures {
		if strings.Contains(extra, sig) {
			status = statusMEV
			break
		}
	}

	return &Result{
		Status: status,
		Reward: rewardResp.Data.Total,
	}, nil
}
