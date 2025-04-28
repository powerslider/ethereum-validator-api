package syncduties

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
	pkgerrors "github.com/pkg/errors"
	"github.com/powerslider/ethereum-validator-api/pkg/beacon"
)

var ErrDutiesNotFound = errors.New("sync duties not found for given epoch")

type Service struct {
	Client       *http.Client
	Cache        *cache.Cache
	ConsensusURL string
}

func NewService(consensusURL string) *Service {
	return &Service{
		ConsensusURL: consensusURL,
		Client:       http.DefaultClient,
		Cache:        cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (s *Service) GetSyncDuties(ctx context.Context, slot uint64) ([]string, error) {
	epoch := slot / 32

	epochStr := strconv.FormatUint(epoch, 10)
	if cached, found := s.Cache.Get(epochStr); found {
		if validators, ok := cached.([]string); ok {
			return validators, nil
		}

		return nil, errors.New("invalid cache type for sync duties")
	}

	validators, err := s.fetchSyncDuties(ctx, epoch)
	if err != nil {
		if !errors.Is(err, ErrDutiesNotFound) {
			return nil, err
		}

		currentEpoch, err := s.GetCurrentEpoch(ctx)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "could not get current epoch after 404")
		}

		validators, err = s.fetchSyncDuties(ctx, currentEpoch)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "fetch duties even after fallback")
		}
	}

	s.Cache.Set(epochStr, validators, 5*time.Minute)

	return validators, nil
}

func (s *Service) GetCurrentEpoch(ctx context.Context) (uint64, error) {
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
		return 0, pkgerrors.Wrap(pkgerrors.Errorf("unexpected status code when fetching head: %d", resp.StatusCode), "fetch head header")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, pkgerrors.Wrap(err, "read head response body")
	}

	var parsed beacon.HeadResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, pkgerrors.Wrap(err, "parse head response")
	}

	epoch := parsed.Data.Header.Message.Slot / 32

	return epoch, nil
}

func (s *Service) fetchSyncDuties(ctx context.Context, epoch uint64) ([]string, error) {
	url := fmt.Sprintf("%s/eth/v1/validator/duties/sync/%d", s.ConsensusURL, epoch)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "create sync duties request")
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "fetch sync duties")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDutiesNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, pkgerrors.Wrap(pkgerrors.Errorf("unexpected status code: %d", resp.StatusCode), "fetch sync duties")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read sync duties response body")
	}

	var parsed beacon.SyncDutiesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, pkgerrors.Wrap(err, "parse sync duties response")
	}

	pubkeys := make([]string, 0, len(parsed.Data))
	for _, d := range parsed.Data {
		pubkeys = append(pubkeys, d.ValidatorPubkey)
	}

	return pubkeys, nil
}
