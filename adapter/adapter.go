package adapter

import . "github.com/rupc/audit/atypes"

type BaasBlockReceiver interface {
	Distill(interface{}) (*DistilledBlock, error)
	// Extract is intended to be used in Distill function
	Extract(interface{}) []*SpecialTransaction
}

type BaasTxSender interface {
	// Tailor transforms a specialtx into Baas-specific transaction format
	Tailor(SpecialTransaction) interface{}
	// Submit sends tailored transaction to Baas endpoint
	Submit(SpecialTransaction) TxID
}

// BaasAdapter lays a bridge between AuditCore and Baas-specific block structures
// i.e., provides an abstraction on heterogeneous Baas-specifics
// BaasAdapter must define behaviors which is different across providers
type BaasAdapter interface {
	BaasBlockReceiver
	BaasTxSender

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

	BaasInitializer
}

// Belows are interface of ledgermgmt from Fabric
// GetBlockchainInfo() (*common.BlockchainInfo, error)
// GetBlockByNumber(blockNumber uint64) (*common.Block, error)
// GetBlocksIterator(startBlockNumber uint64) (ResultsIterator, error)
// GetTransactionByID(txID string) (*peer.ProcessedTransaction, error)
// GetBlockByHash(blockHash []byte) (*common.Block, error)
// GetBlockByTxID(txID string) (*common.Block, error)
// GetTxValidationCodeByTxID(txID string) (peer.TxValidationCode, error)

type BaasInitializer interface {
	Initialize(chainID, blkAddr, txAddr string) error
	// IsValidChainID(chainID string) bool
	// CheckBlockFetchAddress(chainID, endPoint string) bool
	// CheckTxSendAddress(chainID, endPoint string) bool
}
