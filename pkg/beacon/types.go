package beacon

// BlockHeaderResponse is the response from /eth/v1/beacon/headers/{slot}
type BlockHeaderResponse struct {
	Data BlockHeaderData `json:"data"`
}

type BlockHeaderData struct {
	Root      string `json:"root"`
	Header    Header `json:"header"`
	Canonical bool   `json:"canonical"`
}

type Header struct {
	Message   HeaderMessage `json:"message"`
	Signature string        `json:"signature"`
}

type HeaderMessage struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
}

// RewardResponse is the response from /eth/v1/beacon/rewards/blocks/{block_root}
type RewardResponse struct {
	Data                RewardData `json:"data"`
	ExecutionOptimistic bool       `json:"execution_optimistic"`
	Finalized           bool       `json:"finalized"`
}

type RewardData struct {
	ProposerIndex     string `json:"proposer_index"`
	Total             string `json:"total"`
	Attestations      string `json:"attestations"`
	SyncAggregate     string `json:"sync_aggregate"`
	ProposerSlashings string `json:"proposer_slashings"`
	AttesterSlashings string `json:"attester_slashings"`
}

// HeaderResponse is the response from /eth/v1/beacon/headers/{slot}.
type HeaderResponse struct {
	Data HeaderData `json:"data"`
}

type HeaderData struct {
	Root      string      `json:"root"`
	Header    BlockHeader `json:"header"`
	Canonical bool        `json:"canonical"`
}

type BlockHeader struct {
	Signature string       `json:"signature"`
	Message   BlockMessage `json:"message"`
}

type BlockMessage struct {
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
