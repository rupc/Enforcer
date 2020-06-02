package atypes

type BlockSourceAddress string
type BlockSourceStream chan *DistilledBlock
type BlockSourceStreamBufferSize uint32

// TxSubmitPoint is an address of an ordering service node in Baas
type TxSubmitAddress string
type TxID string

// AnalysisRequestChannel accepts start-analysis request from client.
// It does prefix agreement or metric analysis on the given block scope (start, end)
// with a granuality of analysis of window size
type AnalysisRequest struct {
	AnalysisType string `json:"analysis_type"`
	StartBlkNum  string `json:"start_blk_num"`
	EndBlkNum    string `json:"end_blk_num"`
	WindowSize   string `json:"window_size"`
}

type SpecialTransaction struct {
	ChainID   string `json:"chain_id"`
	Number    uint64 `json:"number"`
	Hash      string `json:"hash"`
	Member    string `json:"member"`
	Signature []byte `json:"signature"`
	// Extension MetricExtensions
}

type MetricExtensions struct {
}

// Number: A block height
// ChainID: Name of a blockchain network (e.g., channel name)
// Hash: Datahash of each block (usually included in a block header)
// TxSet: A filtered set of special transactions
// TotalTxNum: For the fractions of special transactions in each block
type DistilledBlock struct {
	Number     uint64                `json:"number"`
	ChainID    string                `json:"chain_id"`
	Hash       string                `json:"hash"`
	PrevHash   string                `json:"prev_hash"`
	TxSet      []*SpecialTransaction `json:"tx_set"`
	TotalTxNum int                   `json:"total_tx_num"`
}

// NewBlockMessage is a message from BaaS Chain, distilled by SDK
// AuditCore delivers NewBlockMessage to AuditorID which matches to (AuditorID, ClientID)
// ClinetID should be a valid owner of the given auditor
type NewBlockMessage struct {
	AuditorID string `json:"auditor_id"`
	ClientID  string `json:"client_id"`

	// Block from BaaS Chain
	Content DistilledBlock `json:"content"`
}

type AuditorOptions struct {
}
