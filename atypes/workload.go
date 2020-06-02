package atypes

import (
	"fmt"
)

type WorkloadGenerator struct {
}

// GetDistilledBlocks_for_window_level_test creates blocks numbered from 0 to size-1
func (w *WorkloadGenerator) GetDistilledBlocks_for_window_level_test(block_size, member_size, tx_size uint64) []*DistilledBlock {
	blocks := make([]*DistilledBlock, block_size)

	members := make([]string, member_size)
	for i := 0; i < int(member_size); i++ {
		members[i] = fmt.Sprintf("%s%d", "client", (i + 1))
	}

	cnt := 0
	memberIndex := 0

	for i := 0; i < int(block_size); i++ {
		var TxSet []*SpecialTransaction

		for j := 0; j < int(tx_size); j++ {

			if memberIndex%int(member_size) == 0 {
				memberIndex = 0
				cnt++
			}

			tx := &SpecialTransaction{
				ChainID: "basschain",
				Number:  uint64(i),
				Hash:    "0xabcdef",
				Member:  members[memberIndex],
			}

			TxSet = append(TxSet, tx)
			memberIndex++

		}

		block := &DistilledBlock{
			Number:   uint64(i + 1),
			ChainID:  "mychannel",
			Hash:     "0xbeef",
			PrevHash: "0xdead",
			TxSet:    TxSet,
		}
		blocks[i] = block
	}

	return blocks
}

// func (w *WorkloadGenerator) GetDistilledBlocks(size uint64) *[]DistilledBlock {

// }
