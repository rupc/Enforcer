package adapter

import . "github.com/rupc/Enforcer/atypes"

type BlockReceiver interface {
	// Distill reduce a block into a minimal form, suit for analysis
	Distill(interface{}) (*DistilledBlock, error)

	// Extract filters special transactions in a distilledBlock
	Extract(interface{}) []*SpecialTransaction
}

type Generator interface {
	// Tailor transforms a specialtx into platform-specific transaction format
	Tailor(SpecialTransaction) interface{}
	// Submit sends tailored transaction to a endpoint (e.g., ordering service in Hyperledger Fabric)
	Submit(SpecialTransaction) TxID
}

// Adapter lays a bridge between AuditCore and platform-specific block structures
// i.e., provides an abstraction on heterogeneous Baas-specifics
// Adapter must define behaviors which is different across providers
type Adapter interface {
	BlockReceiver
	Generator

	// GetNewBlockEventer spawns a worker thread who probes a new block from BlockSourceAddress and
	// injects newly incoming blocks to BlockStream channel.
	// It returns a channel, of which its user(i.e., Auditor) can receive a new block.
	// Expected to be called exactly-once
	//
	// Example Code
	// ---------------------------------------
	// var fa HyperledgerFabricAdapter
	// blkSrc := fa.GetNewBlockListener("orderer1.example.com")
	// while block := <- blkSrc {
	//		...
	//		// Auditor should define a behavior
	//		// when a new block has been arrived from BlockStream channel.
	//		specialTx := audit.Audit(block)
	//		fa.Submit("orderer1.example.com", specialTx)
	//		...
	// }
	// ---------------------------------------
	GetNewBlockListener(chainID, srcAddr string) <-chan *DistilledBlock

	Initializer
}

// Belows are interface of ledgermgmt from Fabric
// GetBlockchainInfo() (*common.BlockchainInfo, error)
// GetBlockByNumber(blockNumber uint64) (*common.Block, error)
// GetBlocksIterator(startBlockNumber uint64) (ResultsIterator, error)
// GetTransactionByID(txID string) (*peer.ProcessedTransaction, error)
// GetBlockByHash(blockHash []byte) (*common.Block, error)
// GetBlockByTxID(txID string) (*common.Block, error)
// GetTxValidationCodeByTxID(txID string) (peer.TxValidationCode, error)

type Initializer interface {
	Initialize(chainID, blkAddr, txAddr string) error
	// IsValidChainID(chainID string) bool
	// CheckBlockFetchAddress(chainID, endPoint string) bool
	// CheckTxSendAddress(chainID, endPoint string) bool
}
