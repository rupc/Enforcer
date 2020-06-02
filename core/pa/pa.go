package pa

import (
	"errors"
	"fmt"

	. "github.com/rupc/Enforcer/atypes"
	"github.com/rupc/Enforcer/logging/flogging"
	"github.com/rupc/Enforcer/util"
)

type PrefixAgreementConfig struct {
	ChainID         string
	NumberOfMembers uint64
	MemberList      []string
	QuorumSize      uint64
}

type PrefixAgreement struct {
	ChainID         string
	NumberOfMembers uint64
	QuorumSize      uint64

	// Exports states
	LatestCommittedHeight          uint64
	LatestPreparedHeight           uint64
	PreviousCommittedHeight        uint64
	PreviousPreparedHeight         uint64
	PreviousHeightWithSelfIssuedTx uint64 // Used to calculate AuditDistance
	UpToDateHeight                 uint64

	// Data structure for consensus
	ConsensusTable *ConsensusTable

	logger *flogging.FabricLogger
}

type PrefixAgreementResult struct {
	Height         uint64 `json:"height"`
	CommitHeight   uint64 `json:"commit_height"`
	PreparedHeight uint64 `json:"prepared_height"`
	HeightGap      uint64 `json:"height_gap"`
}

// ConsensusTable manages all required information for Audit consensus
type ConsensusTable struct {
	// Used when calculating fianl latest committed/prepared height
	// Index implicitly represents height and its value represents aggregated votes
	AggregatedVoteTable []uint64

	// Used when inserting ending mark
	// Intent of this table
	PerMemberHighestVotedBlockTable map[string]uint64

	// Used when interval-aware calculation
	EndingMarkTable map[uint64]string

	// Used only for calculating committed height
	// The value is only added when it lies between g_c and g_p
	CommitBaseTable map[uint64]uint64
}

// 각 멤버 별
type IntervalEntry struct {
	start uint64
	end   uint64
}

func InitializePrefixAgreement(config *PrefixAgreementConfig) *PrefixAgreement {

	consensusTable := initializeConsensusTable()
	pa := &PrefixAgreement{
		ChainID:         config.ChainID,
		NumberOfMembers: config.NumberOfMembers,
		QuorumSize:      util.GetByzantineQuorumSizeFromTotal(config.NumberOfMembers),
		logger:          flogging.MustGetLogger("pa"),
		ConsensusTable:  consensusTable,
	}
	return pa
}

func initializeConsensusTable() *ConsensusTable {
	// TODO Remove fixed size of array. Rather than, use BaseHeight with garbage collecting committed heights
	defaultSize := 20000

	ct := &ConsensusTable{
		AggregatedVoteTable:             make([]uint64, defaultSize),
		PerMemberHighestVotedBlockTable: make(map[string]uint64),
		CommitBaseTable:                 make(map[uint64]uint64),
	}
	return ct
}

func (pa *PrefixAgreement) GetHeights() (uint64, uint64) {
	return pa.LatestPreparedHeight, pa.LatestCommittedHeight
}

func (pa *PrefixAgreement) DoPrefixAgreement(block *DistilledBlock) *PrefixAgreementResult {
	// return when no audit trails are found
	if block.TxSet == nil {
		return &PrefixAgreementResult{
			CommitHeight:   pa.LatestCommittedHeight,
			PreparedHeight: pa.LatestPreparedHeight,
		}
	}

	if pa.UpToDateHeight < block.Number {
		pa.UpToDateHeight = block.Number
	}

	preAggregatedTable, err := pa.ConstructTable(block.TxSet)
	if err != nil {
		pa.logger.Error(err)
		return &PrefixAgreementResult{}
	}

	// debug.PrintStackTrace()

	// Merged AggregatedVoteTable
	mergedTable := pa.Merge(preAggregatedTable)
	p, c := pa.Calculate(mergedTable)

	// Update states
	pa.PreviousPreparedHeight = pa.LatestPreparedHeight
	pa.PreviousCommittedHeight = pa.LatestCommittedHeight

	pa.LatestPreparedHeight = p
	pa.LatestCommittedHeight = c

	// Prepare return results
	r := &PrefixAgreementResult{
		CommitHeight:   c,
		PreparedHeight: p,
		Height:         block.Number,
	}

	pa.logger.Debugf("Calculated prepare[%d], commit[%d] height", p, c)
	return r
}

func (pa *PrefixAgreement) ConstructEndingMarkTable(PerMemberIntervalTable map[string]IntervalEntry) map[uint64]uint64 {
	EndingMarkTable := make(map[uint64]uint64)
	for _, v := range PerMemberIntervalTable {
		EndingMarkTable[v.end]++
	}
	return EndingMarkTable
}

func (pa *PrefixAgreement) ConstructTable(specialTxBuffer []*SpecialTransaction) ([]uint64, error) {
	BaseTable := make(map[uint64]uint64)
	FirstVoteDetectTable := make(map[string]uint64)

	PerMemberIntervalTable := make(map[string]IntervalEntry)
	// PerMemberHighestVotedBlockTable := pa.GetPerMemberHighestVotedBlockTable()

	g_p := pa.LatestPreparedHeight
	g_c := pa.LatestCommittedHeight

	// Highest height in this batch
	var HighestHeight uint64 = 0
	pa.logger.Debugf("g_p[%d], g_c[%d], NumSpecialTx[%d]", g_p, g_c, len(specialTxBuffer))

	// Per special transaction level interpretation
	// Setting up vote tables for efficiently couting votes (i.e., dynamic programming)
	for _, tx := range specialTxBuffer {
		g := uint64(tx.Number)

		err := pa.sanityCheck(tx)
		if err != nil {
			// pa.logger.Warningf(err.Error())
			continue
		}

		// Update commit quorum table based on height, prepared height(g_p), and commit height(g_c)
		// On single pass, guarantees g_c <= g_p
		// while second pass, guarantees g_c < g_p'
		if g_c < g && g <= g_p {
			pa.ConsensusTable.CommitBaseTable[g]++
			pa.logger.Debugf("g[%d]", g)
		} else {
			pa.ConsensusTable.CommitBaseTable[g_p]++
		}

		BaseTable[g]++

		// Ensure that highest vote by a peer is effective in given block
		pa.ensureSingleTxByPeerInBlcok(tx, BaseTable, FirstVoteDetectTable)

		// Ensure that per-member interval is valid
		pa.ensureValidInterval(tx, PerMemberIntervalTable)

		// Update highest g
		if HighestHeight < g {
			HighestHeight = g
		}
	}

	// logger.Notice("for _, tx := range specialTxBuffer { ")
	pa.logger.Debugf("HighestHeight[%d]", HighestHeight)

	if HighestHeight == 0 {
		pa.logger.Warning("HighestHeight:", HighestHeight, "on", pa.ChainID)
		return nil, errors.New("HighestHeight = 0, No special tx in this batch, but how did it reach here?")
	}

	// Block level interpretation
	EndingMarkTable := pa.ConstructEndingMarkTable(PerMemberIntervalTable)
	// logger.Debug(EndingMarkTable)
	PreAggregatedTable := pa.ConstructPreAggregatedTable(HighestHeight, BaseTable, EndingMarkTable)
	// logger.Debug(PreAggregatedTable)

	// Print basetable for debug purpose
	// fmt.Println("BaseTable")
	// for k, v := range BaseTable {
	// fmt.Println("Heihgt", k, "Vote", v)
	// }

	// fmt.Println("EndingMarkTable")
	// for k, v := range EndingMarkTable {
	// fmt.Println(k, v)
	// }

	// fmt.Println("PreAggregatedTable")
	// for k, v := range PreAggregatedTable {
	// fmt.Println(k, v)
	// }

	return PreAggregatedTable, nil
}

func (pa *PrefixAgreement) ConstructPreAggregatedTable(HighestHeight uint64, BaseTable map[uint64]uint64, EndingMarkTable map[uint64]uint64) []uint64 {
	PreAggregatedTable := make([]uint64, HighestHeight+1)
	PreAggregatedTable[HighestHeight] = BaseTable[HighestHeight]

	for i := HighestHeight - 1; i >= 0; i-- {
		if _, ok := BaseTable[i]; !ok {
			BaseTable[i] = 0
		}
		if _, ok := EndingMarkTable[i]; !ok {
			EndingMarkTable[i] = 0
		}

		// EndingMarkTable used to determine when to stop the counting
		PreAggregatedTable[i] = PreAggregatedTable[i+1] + BaseTable[i] - EndingMarkTable[i]
		pa.logger.Debugf("PreAggregatedTable[%d] = PreAggregatedTable[%d] + BaseTable[%d] - EndingMarkTable[%d]", PreAggregatedTable[i], PreAggregatedTable[i+1], BaseTable[i], EndingMarkTable[i])

		if i == 0 {
			// XXX Why 0 next value is executed in this loop?
			// weird...
			pa.logger.Debug("Your are the Weirdo!")
			break
		}
	}

	return PreAggregatedTable
}

func (pa *PrefixAgreement) Merge(PreAggregatedTable []uint64) []uint64 {
	// Add votes with height g in pre-aggregated table to an corresponding entry in aggregated table
	for i := 0; i < len(PreAggregatedTable); i++ {
		pa.ConsensusTable.AggregatedVoteTable[i] += PreAggregatedTable[i]
	}
	return pa.ConsensusTable.AggregatedVoteTable
}

func (pa *PrefixAgreement) Calculate(mergedTable []uint64) (uint64, uint64) {
	// Get previously highest prepared and committed height
	g_p := pa.LatestPreparedHeight
	g_c := pa.LatestCommittedHeight
	q := pa.QuorumSize

	var newPreparedHeight uint64 = 0
	var newCommittedHeight uint64 = 0

	// pa.logger.Notice("[Before]", q, g_p, g_c, len(mergedTable))
	// Find prepared height
	for i := g_p; i < uint64(len(mergedTable)); i++ {
		// pa.logger.Debug("mergedTable[i]", mergedTable[i])
		if mergedTable[i] >= q {
			newPreparedHeight = i
		}
	}

	// Find commit height
	CB := pa.ConsensusTable.CommitBaseTable
	if CB[g_p] >= q {
		newCommittedHeight = g_p
		pa.logger.Debugf("CB[g_p] >= q ; %d >= %d, newCommittedHeight[%d] : ", CB[g_p], q, newCommittedHeight)
	} else if g_p > 1 {
		// Applying table method to calculate a commit height
		for g := g_p - 1; g > g_c; g-- {
			pa.logger.Debugf("CB[%d] = CB[%d] + CB[%d+1] = %d = %d + %d", g, g, g, CB[g], CB[g], CB[g+1])
			CB[g] = CB[g] + CB[g+1]
			if CB[g] >= q {
				newCommittedHeight = g
				break
			}
		}

		if newCommittedHeight == 0 {
			newCommittedHeight = pa.PreviousCommittedHeight
		}

	} else {
		pa.logger.Debug("You are the weirdo!")
	}

	return newPreparedHeight, newCommittedHeight
}

func (pa *PrefixAgreement) VerifyHashConsistency(tx *SpecialTransaction) bool {
	// Removing GetBlockHashByNumber
	// localHash, _ := GetBlockHashByNumber(pa.ChainID, uint64(tx.Number))
	// logger.Noticef("Checking hash conflict with %s on block[%d]:\n(local)%x\n(%s)%x", tx.PeerName, tx.Number, localHash, tx.PeerName, tx.Hash)
	// if !bytes.Equal(tx.Hash, localHash) {
	// return false
	// }
	return true
}

// ensureSingleTxByPeerInBlcok ensures that the highest vote by a peer is effective in a given block,
// For example, if multiple transactions by a peer exist in a block,
// then, choose the highest voted block in this block
// This is due to the fact that chain's linearly expanding prefix
func (pa *PrefixAgreement) ensureSingleTxByPeerInBlcok(tx *SpecialTransaction, BaseTable map[uint64]uint64, FirstVoteDetectTable map[string]uint64) {
	id := tx.Member
	if pa.IsFirstVoteInThisBatch(id, FirstVoteDetectTable) == false {
		previousVotedHeight := FirstVoteDetectTable[id]
		// Why does this related to "already voted" ?
		// logger.Debug(id, " already votes for height ", tx.Number)

		// Decrease a vote with height g' previously voted in this table
		BaseTable[previousVotedHeight]--
	}

	FirstVoteDetectTable[id] = uint64(tx.Number)
}

func (pa *PrefixAgreement) ensureValidInterval(tx *SpecialTransaction, PerMemberIntervalTable map[string]IntervalEntry) {
	id := tx.Member
	height := tx.Number
	if _, ok := PerMemberIntervalTable[id]; !ok {
		// Because height is the recently highest height voted by this peer,
		// and PerMemberHighestVotedBlockTable[id] represents previously highest height,
		// couting from recent highest height to previous highest height is effective
		PerMemberIntervalTable[id] = IntervalEntry{start: height, end: pa.GetPerMemberHighestVotedBlock(id)}
	}

	preservedEnd := PerMemberIntervalTable[id].end
	PerMemberIntervalTable[id] = IntervalEntry{start: height, end: preservedEnd}

	// Update an corresponding entry in PerMemberHighestVotedBlockTable
	pa.ConsensusTable.PerMemberHighestVotedBlockTable[id] = height
}

func (pa *PrefixAgreement) sanityCheck(tx *SpecialTransaction) error {
	id := tx.Member
	height := tx.Number
	g_c := pa.LatestCommittedHeight

	// Verify that genesis block should be ignored
	if height == 0 {
		return errors.New("Genesis block is ignored")
	}

	// Verify that hash is consistent, i.e., comparing local hash vs. voted hash
	// TODO If failed, take some serious actions
	// Comment this when testing this function
	// if !pa.VerifyHashConsistency(tx) {
	//     // DoSomeSeriousAction (e.g., panic)
	//     return errors.New("Confirmed hash conflict![naive fork attack] --> View change on Orderer or Resolution between peers")
	// }

	// Verify that the height of vote is not higher than previous one
	if pa.IsHigherBlock(height, id) == false {
		return errors.New(fmt.Sprintf("Accepts %d's vote on a block only if it is the highest voted block ever by PeerID", id))
	}

	// Verify that height less than commit height is ignored
	if height <= g_c {
		pa.logger.Noticef("height[%d] <= g_c[%d]", height, g_c)
		return errors.New("Blocks below commit height is ignored")
	}

	pa.logger.Debugf("Access block[%d]", height)
	return nil
}
