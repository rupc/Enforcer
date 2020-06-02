package pa

func (pac *PrefixAgreement) GetLatestCommitteHeight() uint64 {
	return pac.LatestCommittedHeight
}

func (pac *PrefixAgreement) GetLatestPreparedHeight() uint64 {
	return pac.LatestPreparedHeight
}

func (pac *PrefixAgreement) GetPreparedHeight() uint64 {
	return pac.LatestPreparedHeight
}

func (pac *PrefixAgreement) GetCommittedHeight() uint64 {
	return pac.LatestCommittedHeight
}

func (pac *PrefixAgreement) SetQuorumSize(q uint64) {
	pac.QuorumSize = q
}

func (pac *PrefixAgreement) GetPerMemberHighestVotedBlockTable() map[string]uint64 {
	return pac.ConsensusTable.PerMemberHighestVotedBlockTable
}

func (pac *PrefixAgreement) GetPerMemberHighestVotedBlock(id string) uint64 {
	return pac.ConsensusTable.PerMemberHighestVotedBlockTable[id]
}

func (pac *PrefixAgreement) IsHigherBlock(height uint64, id string) bool {
	// Empty implies higher block
	if _, ok := pac.ConsensusTable.PerMemberHighestVotedBlockTable[id]; !ok {
		return true
	}

	if pac.ConsensusTable.PerMemberHighestVotedBlockTable[id] > height {
		return false
	}

	return true
}

func (pac *PrefixAgreement) IsFirstVoteInThisBatch(id string, FirstVoteDetectTable map[string]uint64) bool {
	// Already voted in given batch
	if _, ok := FirstVoteDetectTable[id]; ok {
		return false
	}
	return true
}

func (pac *PrefixAgreement) UpdateCorrespondingEntry(id string, height uint64) {
	// tbl := pac.GetPerMemberHighestVotedBlockTable()
	pac.ConsensusTable.PerMemberHighestVotedBlockTable[id] = height
	// tbl[id] = height
}

// func GetBlockHashByNumber(chainID string, height uint64) ([]byte, error) {
//     logger.Debugf("Access block[%d]", height)
//     ledger := peer.GetLedger(chainID)
//     block, err := ledger.GetBlockByNumber(uint64(height) - 1)
//     if err != nil {
//         logger.Error(err)
//         return nil, err
//     }
//     return block.Header.Hash(), err
// }
