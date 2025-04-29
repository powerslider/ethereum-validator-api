package blockreward

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

type Result struct {
	Status string
	Reward string
}

// BlockHeaderResponse is the response from /eth/v1/beacon/headers/{slot}
type BlockHeaderResponse struct {
	Data BlockHeaderData `json:"data"`
}

type BlockHeaderData struct {
	Root      string       `json:"root"`
	Header    BeaconHeader `json:"header"`
	Canonical bool         `json:"canonical"`
}

type BeaconHeader struct {
	Message   BeaconHeaderMessage `json:"message"`
	Signature string              `json:"signature"`
}

type BeaconHeaderMessage struct {
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
