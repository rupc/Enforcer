package audit

import (
	"log"
	"math/rand"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/rupc/Enforcer/atypes"
	"github.com/rupc/Enforcer/core/audit"
	"github.com/stretchr/testify/assert"
)

func getRandomWorkloadSpecialTxes(size int) []*atypes.SpecialTransaction {

	var txes []*atypes.SpecialTransaction

	getRandomSpecialTx := func() *atypes.SpecialTransaction {
		return &atypes.SpecialTransaction{
			ChainID: "basschain",
			Number:  uint64(rand.Intn(100)),
			Hash:    "0xabcdef",
			Member:  "peer" + strconv.Itoa(rand.Intn(4)),
		}
	}

	for i := 0; i < size; i++ {
		tx := getRandomSpecialTx()
		txes = append(txes, tx)
	}

	return txes
}

func getRandomWorkloadBlock(size int) []*atypes.DistilledBlock {
	var blocks []*atypes.DistilledBlock

	for i := 0; i < size; i++ {
		block := &atypes.DistilledBlock{
			Number:   uint64(rand.Intn(100)),
			ChainID:  "auditchain",
			Hash:     "0xabcdef",
			PrevHash: "0xcdefg",
			TxSet:    getRandomWorkloadSpecialTxes(4),
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// BlackBox testing
func TestE2E(t *testing.T) {
	auditor, err := audit.InitializeAuditor("auditor1", "client1", "basschain", "auditchain", "")

	logger := flogging.MustGetLogger("audit_test")
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Init auditor")
	_ = auditor

	// blocksWorkload := getRandomWorkloadBlock(10)
	blocksWorkload := getWorkloadForCorrectness()
	// blockIndex := 0
	// time.Sleep(time.Second * 1)

	auditor.NewBlockArrivalEvent(blocksWorkload[0])
	s := <-auditor.StateOutputChannel
	logger.Debug("Output: %+v", s)
	assert.Equal(t, uint64(1), uint64(s.PrepareHeight), "They should be equal")

	auditor.NewBlockArrivalEvent(blocksWorkload[1])
	s = <-auditor.StateOutputChannel
	logger.Debug("Output: %+v", s)
	assert.Equal(t, uint64(4), uint64(s.PrepareHeight), "They should be equal")

	// ch := make(chan int)
	// <-ch
}

// This test is carefully chosen workloads to test PrefixAgreement module
func getWorkloadForCorrectness() []*atypes.DistilledBlock {
	var workload []*atypes.DistilledBlock

	// Phase 1 workload
	tx1 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  1,
		Hash:    "0xabcdef",
		Member:  "peer1",
	}
	tx2 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  3,
		Hash:    "0xabcdef",
		Member:  "peer2",
	}
	tx3 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  0,
		Hash:    "0xabcdef",
		Member:  "peer3",
	}
	tx4 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  2,
		Hash:    "0xabcdef",
		Member:  "peer4",
	}

	// fmt.Println("g_p", p, "g_c", c)
	// assert.Equal(t, uint64(1), uint64(p), "They should be equal")
	txes1 := []*atypes.SpecialTransaction{tx1, tx2, tx3, tx4}

	block1 := &atypes.DistilledBlock{
		Number:   1,
		ChainID:  "auditchain",
		Hash:     "0xbbbb",
		PrevHash: "0xdddd",
		TxSet:    txes1,
	}

	// Phase 2 workload
	tx5 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  4,
		Hash:    "0xabcdef",
		Member:  "peer1",
	}
	tx6 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  5,
		Hash:    "0xabcdef",
		Member:  "peer2",
	}
	tx7 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  4,
		Hash:    "0xabcdef",
		Member:  "peer3",
	}
	tx8 := &atypes.SpecialTransaction{
		ChainID: "basschain",
		Number:  6,
		Hash:    "0xabcdef",
		Member:  "peer4",
	}

	txes2 := []*atypes.SpecialTransaction{tx5, tx6, tx7, tx8}

	block2 := &atypes.DistilledBlock{
		Number:   2,
		ChainID:  "auditchain",
		Hash:     "0xeeee",
		PrevHash: "0xbbbb",
		TxSet:    txes2,
	}

	// fmt.Println("g_p", p2, "g_c", c2)
	// assert.Equal(t, uint64(4), uint64(p2), "They should be equal")

	workload = append(workload, block1)
	workload = append(workload, block2)

	return workload
}
