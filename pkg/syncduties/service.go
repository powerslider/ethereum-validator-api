package syncduties

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrDutiesNotFound     = errors.New("sync duties not found for given slot")
	ErrSlotTooFarInFuture = errors.New("slot is too far in the future")
	ErrSlotWasMissed      = errors.New("slot was missed")
)

type Service struct {
	Client       *http.Client
	ConsensusURL string
}

func NewService(consensusURL string) *Service {
	return &Service{
		ConsensusURL: consensusURL,
		Client:       http.DefaultClient,
	}
}

func (s *Service) GetSyncDuties(ctx context.Context, slot uint64) ([]string, error) {
	currentSlot, err := s.getCurrentSlot(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch current slot")
	}

	if slot > currentSlot+1 {
		return nil, ErrSlotTooFarInFuture
	}

	validatorIndexes, err := s.fetchSyncCommitteeIndexes(ctx, slot)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch sync committee")
	}

	validators, err := s.fetchValidatorsByIDs(ctx, slot, validatorIndexes)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch validators by ids")
	}

	return validators, nil
}

func (s *Service) getCurrentSlot(ctx context.Context) (uint64, error) {
	url := s.ConsensusURL + "/eth/v1/beacon/headers/head"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "create head request")
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "fetch head header")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, pkgerrors.Wrap(
			pkgerrors.Errorf("unexpected status code when fetching head: %d", resp.StatusCode),
			"fetch head header")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "read head response body")
	}

	var parsed beacon.HeadResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, pkgerrors.Wrap(err, "parse head response")
	}

	return parsed.Data.Header.Message.Slot, nil
}

func (s *Service) fetchSyncCommitteeIndexes(ctx context.Context, slot uint64) ([]string, error) {
	url := fmt.Sprintf("%s/eth/v1/beacon/states/%d/sync_committees", s.ConsensusURL, slot)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "create sync committee request")
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "execute sync committee request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read sync committee response body")
	}

	if resp.StatusCode != http.StatusOK {
		var syncCommitteeErr beacon.SyncCommitteeError
		if err = json.Unmarshal(body, &syncCommitteeErr); err != nil {
			return nil, pkgerrors.Wrap(err, "parse sync committee error")
		}

		if resp.StatusCode == http.StatusNotFound {
			return nil, pkgerrors.Wrap(ErrDutiesNotFound, syncCommitteeErr.Message)
		}

		slotWasMissed := syncCommitteeErr.Code == http.StatusBadRequest &&
			strings.Contains(syncCommitteeErr.Message, "is not activated for Altair")
		if slotWasMissed {
			return nil, pkgerrors.Wrap(ErrSlotWasMissed, syncCommitteeErr.Message)
		}
	}

	var parsed beacon.SyncCommitteeResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return nil, pkgerrors.Wrap(err, "parse sync committee response")
	}

	return parsed.Data.Validators, nil
}

func (s *Service) fetchValidatorsByIDs(ctx context.Context, slot uint64, ids []string) ([]string, error) {
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

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "execute validator chunk request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, pkgerrors.Wrap(
			pkgerrors.Errorf("unexpected status code: %d", resp.StatusCode),
			"fetch validator chunk response")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read validator response body")
	}

	var parsed beacon.ValidatorResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, pkgerrors.Wrap(err, "parse validator response")
	}

	var validators []string
	for _, v := range parsed.Data {
		validators = append(validators, v.Validator.Pubkey)
	}

	return validators, nil
}
