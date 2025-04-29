package blockreward

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	statusVanilla = "vanilla"
	statusMEV     = "mev"
)

var (
	ErrSlotMissedOrDoesNotExist = errors.New("slot was missed or does not exist")
	ErrSlotInFuture             = errors.New("slot is in the future")
)

// Service provides block reward calculation functionality.
type Service struct {
	ExecClient      *ethclient.Client
	ConsensusClient *http.Client
	ConsensusURL    string
}

// NewService creates a new block reward service instance.
func NewService(client *ethclient.Client, consensusURL string) *Service {
	return &Service{
		ExecClient:      client,
		ConsensusClient: http.DefaultClient,
		ConsensusURL:    consensusURL,
	}
}

// GetBlockReward calculates the block reward earned by the validator at a given slot.
// It returns the block status ("vanilla" or "mev") and the reward amount in Gwei.
func (s *Service) GetBlockReward(ctx context.Context, slot uint64) (*Result, error) {
	// Step 1: Get block header to retrieve proposer index and block root
	headerResp, err := s.getBeaconHeader(ctx, slot)
	if err != nil {
		return nil, err
	}

	blockRoot := headerResp.Data.Root

	// Step 2: Get consensus-layer reward (already in Gwei)
	rewardResp, err := s.getBlockRewardFromConsensus(ctx, blockRoot)
	if err != nil {
		return nil, err
	}

	// Step 3: Fetch execution block to inspect ExtraData for MEV tag
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

func (s *Service) getBeaconHeader(ctx context.Context, slot uint64) (*BlockHeaderResponse, error) {
	url := fmt.Sprintf("%s/eth/v1/beacon/headers/%d", s.ConsensusURL, slot)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, pkgerrors.Wrap(err, "create beacon header request")
	}

	resp, err := s.ConsensusClient.Do(req)

	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch beacon header")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read beacon header response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleBeaconAPIError(body, resp.StatusCode)
	}

	var out BlockHeaderResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, pkgerrors.Wrap(err, "parse beacon header response")
	}

	return &out, nil
}

func (s *Service) getBlockRewardFromConsensus(ctx context.Context, blockRoot string) (*RewardResponse, error) {
	url := fmt.Sprintf("%s/eth/v1/beacon/rewards/blocks/%s", s.ConsensusURL, blockRoot)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "create block reward request")
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.ConsensusClient.Do(req)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch block reward")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read block reward response body")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSlotMissedOrDoesNotExist
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleBeaconAPIError(body, resp.StatusCode)
	}

	var out RewardResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, pkgerrors.Wrap(err, "parse block reward response")
	}

	return &out, nil
}

func handleBeaconAPIError(body []byte, statusCode int) error {
	var apiError beacon.APIError
	if err := json.Unmarshal(body, &apiError); err != nil {
		return pkgerrors.Wrap(err, "parse beacon header error")
	}

	if statusCode == http.StatusNotFound {
		return pkgerrors.Wrap(ErrSlotMissedOrDoesNotExist, apiError.Message)
	}

	return pkgerrors.Wrap(beacon.ErrUnexpectedStatusCode(statusCode), apiError.Message)
}
