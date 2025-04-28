package blockreward

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	pkgerrors "github.com/pkg/errors"
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

// Service provides block reward calculation functionality.
type Service struct {
	Client *ethclient.Client
}

type Result struct {
	Status string
	Reward string
}

// NewService creates a new block reward service instance.
func NewService(client *ethclient.Client) *Service {
	return &Service{
		Client: client,
	}
}

// GetBlockReward calculates the block reward earned by the validator at a given slot.
// It returns the block status ("vanilla" or "mev") and the reward amount in Gwei.
func (s *Service) GetBlockReward(ctx context.Context, slot uint64) (*Result, error) {
	block, err := s.Client.BlockByNumber(ctx, new(big.Int).SetUint64(slot))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch block by number")
	}

	coinbase := block.Coinbase()

	balanceBefore, err := s.Client.BalanceAt(ctx, coinbase, new(big.Int).SetUint64(slot-1))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch balance before block")
	}

	balanceAfter, err := s.Client.BalanceAt(ctx, coinbase, new(big.Int).SetUint64(slot))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch balance after block")
	}

	reward := new(big.Int).Sub(balanceAfter, balanceBefore)
	rewardInGWEI := new(big.Int).Div(reward, big.NewInt(1e9))

	extra := strings.ToLower(strings.TrimSpace(string(block.Extra())))
	status := statusVanilla

	for _, sig := range mevRelaySignatures {
		if strings.Contains(extra, sig) {
			status = statusMEV
			break
		}
	}

	return &Result{
		Status: status,
		Reward: rewardInGWEI.String(),
	}, nil
}
