package ibm

import (
	. "github.com/rupc/Enforcer/atypes"
)

type Provider struct {
}

func (p *Provider) Initialize(chainID, blockFetchAddress, txSendAddress string) error {
	return nil
}
func (p *Provider) Distill(block interface{}) (*DistilledBlock, error) {
	return nil, nil
}

func (p *Provider) Extract(block interface{}) []*SpecialTransaction {
	return nil
}

func (p *Provider) Tailor(tx SpecialTransaction) interface{} {
	return nil
}

func (p *Provider) Submit(tx SpecialTransaction) TxID {
	return ""
}

func (p *Provider) GetNewBlockListener(chainID, srcAddr string) <-chan *DistilledBlock {
	return nil
}
