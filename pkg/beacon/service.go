package beacon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrSlotMissedOrDoesNotExist = errors.New("slot was missed or does not exist")
	ErrSlotInFuture             = errors.New("slot is in the future")
	ErrDutiesNotFound           = errors.New("sync duties not found for given slot")
	ErrSlotWasMissed            = errors.New("slot was missed")
)

// Service provides a way to interact with the consensus layer.
type Service struct {
	ConsensusClient *http.Client
	ConsensusURL    string
}

// NewService creates a new beacon service instance for interacting with the consensus layer.
func NewService(consensusURL string) *Service {
	return &Service{
		ConsensusClient: http.DefaultClient,
		ConsensusURL:    consensusURL,
	}
}

// GetBeaconHeader retrieves the beacon block header for a specific slot.
// Returns ErrSlotInFuture or ErrSlotMissedOrDoesNotExist when appropriate.
func (s *Service) GetBeaconHeader(ctx context.Context, slot uint64) (*BlockHeaderResponse, error) {
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

// GetBlockRewardFromConsensus retrieves the block reward breakdown (in Gwei)
// from the consensus layer using the block root.
// Returns ErrSlotInFuture or ErrSlotMissedOrDoesNotExist when applicable.
func (s *Service) GetBlockRewardFromConsensus(ctx context.Context, blockRoot string) (*RewardResponse, error) {
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

	if resp.StatusCode != http.StatusOK {
		return nil, handleBeaconAPIError(body, resp.StatusCode)
	}

	var out RewardResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, pkgerrors.Wrap(err, "parse block reward response")
	}

	return &out, nil
}

// GetCurrentSlot fetches the current head slot from the consensus layer.
func (s *Service) GetCurrentSlot(ctx context.Context) (uint64, error) {
	url := s.ConsensusURL + "/eth/v1/beacon/headers/head"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "create head request")
	}

	resp, err := s.ConsensusClient.Do(req)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "fetch head header")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, pkgerrors.Wrap(
			ErrUnexpectedStatusCode(resp.StatusCode),
			"fetch head header")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "read head response body")
	}

	var parsed HeaderResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, pkgerrors.Wrap(err, "parse head response")
	}

	return parsed.Data.Header.Message.Slot, nil
}

// FetchSyncCommitteeIndexes retrieves the list of validator indices
// assigned to sync committee duties for a given slot.
func (s *Service) FetchSyncCommitteeIndexes(ctx context.Context, slot uint64) ([]string, error) {
	url := fmt.Sprintf("%s/eth/v1/beacon/states/%d/sync_committees", s.ConsensusURL, slot)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "create sync committee request")
	}

	resp, err := s.ConsensusClient.Do(req)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "execute sync committee request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read sync committee response body")
	}

	if resp.StatusCode != http.StatusOK {
		var apiError APIError
		if err = json.Unmarshal(body, &apiError); err != nil {
			return nil, pkgerrors.Wrap(err, "parse sync committee error")
		}

		if resp.StatusCode == http.StatusNotFound {
			return nil, pkgerrors.Wrap(ErrDutiesNotFound, apiError.Message)
		}

		slotWasMissed := apiError.Code == http.StatusBadRequest &&
			strings.Contains(apiError.Message, "is not activated for Altair")
		if slotWasMissed {
			return nil, pkgerrors.Wrap(ErrSlotWasMissed, apiError.Message)
		}

		return nil, pkgerrors.Wrap(ErrUnexpectedStatusCode(resp.StatusCode), apiError.Message)
	}

	var parsed SyncCommitteeResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, pkgerrors.Wrap(err, "parse sync committee response")
	}

	return parsed.Data.Validators, nil
}

// FetchValidatorsByIDs fetches the validator public keys for a list of indices
// at a specific slot using concurrent chunked requests.
func (s *Service) FetchValidatorsByIDs(ctx context.Context, slot uint64, ids []string) ([]string, error) {
	var (
		result    []string
		chunkSize = 100
		mu        = &sync.Mutex{} // to safely append to shared slice
	)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < len(ids); i += chunkSize {
		start := i
		end := i + chunkSize

		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[start:end]
		chunk = append([]string(nil), chunk...) // safe copy

		g.Go(func() error {
			validators, err := s.fetchValidatorChunk(ctx, slot, chunk)
			if err != nil {
				return err
			}

			mu.Lock()
			result = append(result, validators...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, pkgerrors.Wrap(err, "fetch validator chunk")
	}

	return result, nil
}

// fetchValidatorChunk fetches a single batch of validators for the given IDs at a slot.
func (s *Service) fetchValidatorChunk(ctx context.Context, slot uint64, ids []string) ([]string, error) {
	url := fmt.Sprintf("%s/eth/v1/beacon/states/%d/validators?", s.ConsensusURL, slot)

	for i, id := range ids {
		if i > 0 {
			url += "&"
		}

		url += "id=" + id
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "create validator chunk request")
	}

	resp, err := s.ConsensusClient.Do(req)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "execute validator chunk request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, pkgerrors.Wrap(
			ErrUnexpectedStatusCode(resp.StatusCode),
			"fetch validator chunk response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read validator response body")
	}

	var parsed ValidatorListResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, pkgerrors.Wrap(err, "parse validator response")
	}

	var validators []string
	for _, v := range parsed.Data {
		validators = append(validators, v.Validator.Pubkey)
	}

	return validators, nil
}

// handleBeaconAPIError parses a consensus-layer error response and returns a typed Go error
// such as ErrSlotInFuture, ErrSlotMissedOrDoesNotExist, or a generic status code error.
func handleBeaconAPIError(body []byte, statusCode int) error {
	var apiError APIError
	if err := json.Unmarshal(body, &apiError); err != nil {
		return pkgerrors.Wrap(err, "parse beacon header error")
	}

	if statusCode == http.StatusNotFound {
		return pkgerrors.Wrap(ErrSlotMissedOrDoesNotExist, apiError.Message)
	}

	return pkgerrors.Wrap(ErrUnexpectedStatusCode(statusCode), apiError.Message)
}
