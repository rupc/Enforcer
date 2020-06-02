package provider

//go:generate stringer -type=BaaSProvider
type BaaSProvider int

const (
	HyperledgerFabric BaaSProvider = iota
	Kaleido
	AmazonManagedBlockchain
	HyperledgerCello
	IBM
	Microsoft
	SAP
	ShieldCure
	Blocko
	HyperledgerIroha
	HyperledgerIndy
	HyperledgerSawtooth
)

// type BlockType_HyperledgerFabric = common.Block
// type BlockType_Kaleido = adapter.Block
// type BlockType_AmazonManagedBlockchain = adapter.Block
// type BlockType_HyperledgerCello = adapter.Block
// type BlockType_IBM = adapter.Block
// type BlockType_Microsoft = adapter.Block
// type BlockType_SAP = adapter.Block
// type BlockType_ShieldCure = adapter.Block
// type BlockType_Blocko = adapter.Block
// type BlockType_HyperledgerIroha = adapter.Block
