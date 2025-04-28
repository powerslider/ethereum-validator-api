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

type SyncDutiesResponse struct {
	Data []SyncDuty `json:"data"`
}

type SyncDuty struct {
	ValidatorPubkey string `json:"validator_pubkey"`
}
