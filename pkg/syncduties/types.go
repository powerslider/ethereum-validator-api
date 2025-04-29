package syncduties

// BeaconHeaderResponse is the response from /eth/v1/beacon/headers/{slot}.
type BeaconHeaderResponse struct {
	Data BeaconHeaderData `json:"data"`
}

type BeaconHeaderData struct {
	Root      string            `json:"root"`
	Header    BeaconBlockHeader `json:"header"`
	Canonical bool              `json:"canonical"`
}

type BeaconBlockHeader struct {
	Signature string             `json:"signature"`
	Message   BeaconBlockMessage `json:"message"`
}

type BeaconBlockMessage struct {
	ProposerIndex string `json:"proposer_index"`
	Slot          uint64 `json:"slot,string"`
}

// ValidatorListResponse is the response from /eth/v1/beacon/states/{state}/validators.
type ValidatorListResponse struct {
	Data []ValidatorEntry `json:"data"`
}

type ValidatorEntry struct {
	Index     string        `json:"index"`
	Validator ValidatorInfo `json:"validator"`
}

type ValidatorInfo struct {
	Pubkey                string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
}

// SyncCommitteeResponse is the response from /eth/v1/beacon/states/{state}/sync_committees.
type SyncCommitteeResponse struct {
	Data SyncCommitteeData `json:"data"`
}

type SyncCommitteeData struct {
	Validators []string `json:"validators"`
}
