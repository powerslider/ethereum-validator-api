package beacon

// HeadResponse represents the response from /eth/v1/beacon/headers/head.
type HeadResponse struct {
	Data HeadData `json:"data"`
}

type HeadData struct {
	Header Header `json:"header"`
}

type Header struct {
	Message Message `json:"message"`
}

type Message struct {
	Slot uint64 `json:"slot,string"`
}

type ValidatorResponse struct {
	Data []ValidatorData `json:"data"`
}

type ValidatorData struct {
	Index     string        `json:"index"`
	Validator ValidatorInfo `json:"validator"`
}

type ValidatorInfo struct {
	Pubkey string `json:"pubkey"`
}

// SyncCommitteeResponse defines the parsed structure for sync committee duties.
type SyncCommitteeResponse struct {
	Data SyncCommitteeData `json:"data"`
}

type SyncCommitteeData struct {
	Validators []string `json:"validators"`
}

type SyncCommitteeError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
