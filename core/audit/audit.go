package audit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-redis/redis"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/rupc/audit/adapter"
	"github.com/rupc/audit/adapter/provider/fabric_client"
	"github.com/rupc/audit/adapter/provider/ibm"
	"github.com/rupc/audit/adapter/provider/kaleido"
	"github.com/rupc/audit/adapter/provider/microsoft"
	"github.com/rupc/audit/adapter/provider/oracle"
	"github.com/rupc/audit/adapter/provider/sap"
	"github.com/rupc/audit/adapter/provider/vmware"
	"github.com/rupc/audit/atypes"
	"github.com/rupc/audit/core"
	"github.com/rupc/audit/core/fetcher"
	"github.com/rupc/audit/core/pa"
)

var logger *flogging.FabricLogger

func init() {
	logger = flogging.MustGetLogger("audit")
}

type State struct {
	PrepareHeight uint64  `json:"prepare_height"`
	CommitHeight  uint64  `json:"commit_height"`
	ClientID      string  `json:"client_id"`
	AuditorID     string  `json:"auditor_id"`
	Score         float64 `json:"score"`
}

type GeneratorPolicy struct {
	// Expected special transactions per blocks, i.e., audit fidelity
	ExpectedInterBlockDistance uint64
}

// type CryptoKeyPair struct {
//     PublicKey  string
//     PrivateKey string
// }

// func CreateKeyPair() {
//     privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
//     if err != nil {
//         panic(err)
//     }
//     publicKey := privateKey.PublicKey

//     msg := "hello, world"
//     hash := sha256.Sum256([]byte(msg))

//     r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
//     if err != nil {
//         panic(err)
//     }
//     fmt.Printf("signature: (0x%x, 0x%x)\n", r, s)

//     valid := ecdsa.Verify(&privateKey.PublicKey, hash[:], r, s)
//     fmt.Println("signature verified:", valid)
// }

type Auditor struct {
	// Client identity that requests to create the auditor
	CreatorID string

	// User-provided audit instance name
	AuditorID string

	// RecordChainID, is i.e., AuditChain
	RecordChainID string

	// TargetChain, is i.e., BaaSChain
	TargetChainID string

	// BlockArrivalChannel receives a new block from SDK server
	// Auditor synchronosly waits new block from this channel.
	// SpecialTxGenerator must be called for this
	BlockArrivalChannel chan *atypes.DistilledBlock

	// StateOutputChannel exports internal events such as prepare/commit height to caller.
	// Reason to require this is that an external user need to know what's going on in the Auditor
	// For example, in Go blockbox test, test function need internal consensus values,
	// but it cannot access the internal data structures.
	StateOutputChannel chan *State

	// Fidelity is a period that an auditor generates a special transaction
	Fidelity string

	// TxSubmitAddress is an IP address where auditors send txes
	TxSubmitAddress    string
	StateSendAddress   string
	MetricFetchAddress string
	Provider           adapter.BaasAdapter

	pa *pa.PrefixAgreement
	ma *core.MetricAnalysis

	// Redis DB client for fetching blocks and generating events (number, prepared, commit)
	r *redis.Client

	// SentBlock contains paris of (blockNum, sent-or-not field)
	// Because there should be exactly one special transaction per block,
	// an auditor need to main block numbers that he sent before
	// XXX Underlying dtat structure would be bitmap-like
	SentBlocks map[uint64]bool

	// InterpretedBlocks has same purpose as SentBlokcs but with RecordChain
	// An anditor should interpret exactly one block from AuditChain
	// XXX Underlying dtat structure would be bitmap-like
	InterpretedBlocks map[uint64]bool

	Fetcher fetcher.Fetcher

	Policy  GeneratorPolicy
	Version string
	logger  *flogging.FabricLogger
}

// type DoAudit interface {
//     CreateSpecialTransaction(block *atypes.DistilledBlock) *atypes.SpecialTransaction
//     Submit(tx *atypes.SpecialTransaction)
//     Interpret(block *atypes.DistilledBlock) *atypes.SpecialTransaction
//     Run()
// }

var auditorPools = struct {
	m    sync.RWMutex
	list map[string]*Auditor
}{list: make(map[string]*Auditor)}

func GetAuditorIDs() []string {
	auditors := make([]string, 0, len(auditorPools.list))
	for k := range auditorPools.list {
		auditors = append(auditors, k)
	}
	return auditors
}

func GetAuditorByID(id string) *Auditor {
	auditorPools.m.Lock()
	defer auditorPools.m.Unlock()
	return auditorPools.list[id]
}

func IsValidAuditor(id string) bool {
	auditorPools.m.Lock()
	defer auditorPools.m.Unlock()
	if _, ok := auditorPools.list[id]; ok {
		return true
	}
	return false
}

// InitializeAuditor is top-level initializer of Audit which returns AuditInstacne
// It prepares all the data structures needed to start auditor service
// It registers current auditor instance to audit_mgmt
func InitializeAuditor(auditorID, creatorID, target, record, txsubmit_address, fidelity, state_send_address, fetch_method, metric_fetch_address, report_address, fetch_address string, MetricNames []string) (*Auditor, error) {
	auditorPools.m.Lock()
	for k, v := range auditorPools.list {
		logger.Debugf("auditorPools[%s] = [%s]", k, v)
	}

	if _, ok := auditorPools.list[auditorID]; ok {

		errMsg := fmt.Sprintf("%s already exists, size of pools[%d]", auditorID, len(auditorPools.list))
		// logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	auditorPools.m.Unlock()

	// XXX For now, NumberOfMembers is: 4, after finishing testing
	paConfig := pa.PrefixAgreementConfig{
		ChainID:         target,
		NumberOfMembers: 4,
	}
	maxAnalysisRequestBufferSize := 100
	maxBlockArrivalBufferSize := 100
	maxStateOutBufferSize := 100

	// XXX For now, Period 10 seconds, but it will be best-fit number
	maConfig := core.MetricAnalysisConfig{
		FetchMethod:            fetch_method,
		FetchAddress:           metric_fetch_address,
		ReportAddress:          report_address,
		Period:                 10,
		ScriptConfigFile:       "scripts/config.json",
		AnalysisRequestChannel: make(chan *atypes.AnalysisRequest, maxAnalysisRequestBufferSize),
	}

	pa := pa.InitializePrefixAgreement(paConfig)
	logger.Debug("PrefixAgreement initialized")

	ma := core.InitializeMetricAnalysis(maConfig)
	logger.Debug("MetricAnalysis initialized")

	// p, err := GetProviderFromPlatform(platform, chainID, blkAddr, txAddr)

	logger := flogging.MustGetLogger(auditorID)

	if txsubmit_address == "" {
		logger.Warn("TxSubmitAddress is empty, its should be test-only purpose")
	}

	policy := GeneratorPolicy{
		ExpectedInterBlockDistance: 1,
	}

	BlockArrivalChannel := make(chan *atypes.DistilledBlock, maxBlockArrivalBufferSize)

	// get redis client
	r, err := GetRedisDBClient(fetch_address)
	if err != nil {
		logger.Error("Cannot get a redis client for", fetch_address)
		panic(err)
	}

	Fetcher := fetcher.Fetcher{}
	Fetcher.Initialize(r)

	auditor := &Auditor{
		AuditorID:           auditorID,
		CreatorID:           creatorID,
		TargetChainID:       target,
		RecordChainID:       record,
		BlockArrivalChannel: BlockArrivalChannel,
		StateOutputChannel:  make(chan *State, maxStateOutBufferSize),
		Fidelity:            fidelity,
		// Provider:  p,
		TxSubmitAddress:    txsubmit_address,
		StateSendAddress:   state_send_address,
		MetricFetchAddress: metric_fetch_address,
		logger:             flogging.MustGetLogger(auditorID),
		// Version:   version,
		pa:                pa,
		ma:                ma,
		r:                 r,
		SentBlocks:        make(map[uint64]bool),
		InterpretedBlocks: make(map[uint64]bool),
		Policy:            policy,
		Fetcher:           Fetcher,
	}

	// Add a new auditor to auditorPools
	auditorPools.m.Lock()
	auditorPools.list[auditorID] = auditor
	auditorPools.m.Unlock()

	// If dev mode, just run on initialization, otherwise, it should run on explicit run request

	// Spawn a new auditor goroutine
	auditor.Run()

	return auditor, nil
}

func (a *Auditor) SpecialTransactionGenerator(block *atypes.DistilledBlock) {
}

func (a *Auditor) NewBlockArrivalEvent(block *atypes.DistilledBlock) {
	// It should be non blocking
	a.BlockArrivalChannel <- block
}

func GetProviderFromPlatform(platform, chainID, blkAddr, txAddr string) (adapter.BaasAdapter, error) {
	var provider adapter.BaasAdapter
	switch platform {
	// case "hyperledger/fabric":
	//     provider = &fabric.Provider{}
	case "hyperledger/fabric-client":
		provider = &fabric_client.Provider{}
	case "kaleido":
		provider = &kaleido.Provider{}
	case "ibm":
		provider = &ibm.Provider{}
	case "microsoft":
		provider = &microsoft.Provider{}
	case "oracle":
		provider = &oracle.Provider{}
	case "vmware":
		provider = &vmware.Provider{}
	case "sap":
		provider = &sap.Provider{}
	default:
		_ = provider
		return provider, errors.New(fmt.Sprintf("Invalid platform: %s", platform))
	}

	err := provider.Initialize(chainID, blkAddr, txAddr)
	if err != nil {
		return provider, err
	}
	return provider, nil
}

func (a *Auditor) Generator(block *atypes.DistilledBlock) {
	if block.Number%a.Policy.ExpectedInterBlockDistance == 0 {
		tx := a.createSpecialTransaction(block)
		a.submit(tx)
	}
}

func (a *Auditor) Run() {
	a.logger.Infof("Auditor[%s] runs on target[%s] and record[%s]", a.AuditorID, a.TargetChainID, a.RecordChainID)

	// Fetcher subscriber has its own goroutine
	a.logger.Infof("Start new block subscriber !")
	a.Fetcher.StartNewBlockSubscriber(a.BlockArrivalChannel)

	go func() {
		for {
			block := <-a.BlockArrivalChannel
			a.logger.Infof("Auditor[%s] received a new block[%d], target[%s], record[%s], numOfTx[%d]", a.AuditorID, block.Number, a.TargetChainID, a.RecordChainID, len(block.TxSet))

			var result *core.InterpretResult

			// Note that if TargetChainID == RecordChainID,
			// then the system follows one-chain approach(on-chain auditing)
			// Otherwise, it follows two-chain approach(off-chain auditing)
			switch block.ChainID {
			case a.TargetChainID:
				// if block was made before, then ignore
				if a.SentBefore(block.Number) {
					a.logger.Debugf("Special tx for block[%d] was made before", block.Number)
					continue
				}
				a.SentBlocks[block.Number] = true
				// a.Generator(block)
				fallthrough
			case a.RecordChainID:
				if a.InterpretedBefore(block.Number) {
					a.logger.Debugf("Block[%d] was interpreted before", block.Number)
					continue
				}
				a.InterpretedBlocks[block.Number] = true

				result = a.interpret(block)
				a.consumeResult(result)
			default:
				a.logger.Infof("No matching chain")
				continue
			}
			// a.logger.Info("%+v, %+v", pa_result, ma_result)

			// Send result to StateOutputChannel
		}
	}()

	// XXX Note that metric analysis is implemented separately from prefix agreement
	// This is because prefix agreement is an on-line algorithm,
	// while metric analysis is an off-line algorithm
	// However, both prefix agreement and metric analaysis can be invoked by request of start analysis

	// Offline metric analysis algorithm
	// Run metric analysis thread separately
	// go a.ma.Run()

	// Run state sender thread
	// go a.RunSenderThread()

}

// !!!!!!!!!!!!!!!NOTE!!!!!!!!!!!!!!!
// SenderThread should be deprecated after I start to use redis as a communication medium...
// This thread was originally intended to send internal data to a node client server
// func (a *Auditor) RunSenderThread() {
//     for {
//         stateOutput := <-a.StateOutputChannel
//         if stateOutput == nil {
//             continue
//         }

//         if a.StateSendAddress == "" {
//             logger.Debug("Received a stateOutput, but StateSendAddress is empty")
//             continue
//         }

//         // Send stateOutput to Client
//         jsonValue, err := json.Marshal(*stateOutput)
//         resp, err := http.Post(a.StateSendAddress, "application/json", bytes.NewBuffer(jsonValue))
//         if err != nil {
//             a.logger.Infof("Sending stateOutput failed", err)
//         }
//         a.logger.Infof(resp.Status)
//     }
// }

// End-user or application might use interpretation result to detect anomalous faults
// consumeResult fnctions
func (a *Auditor) consumeResult(res *core.InterpretResult) {
	// Send interpretation result to Client...
	// state := &State{
	// PrepareHeight: res.PAResult.PreparedHeight,
	// CommitHeight:  res.PAResult.CommitHeight,
	// ClientID:      a.CreatorID,
	// AuditorID:     a.AuditorID,
	// }

	// a.logger.Infof("ConsumeResult: %+v", state)

	bytes, err := json.Marshal(res.PAResult)
	if err != nil {
		a.logger.Error("cannot marshal !")
	}

	// publish output of a prefix analysis
	a.r.Publish("prefix analysis", bytes)

	// a.StateOutputChannel <- state
}

func (a *Auditor) submit(tx *atypes.SpecialTransaction) {
	// TxSubmitAddress == "" implies testing purposes
	if a.TxSubmitAddress == "" || a.TxSubmitAddress == "test" {
		logger.Debug("Received, but TxSubmitAddress is empty!")
		return
	}

	a.logger.Debug("Send to tx.TxSubmitAddress")

	go func() {
		// send http request to sendAddr (maybe node SDK)
		jsonValue, err := json.Marshal(*tx)
		resp, err := http.Post(a.TxSubmitAddress, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			a.logger.Infof("Sending special tx failed", err)
		}
		a.logger.Infof(resp.Status)
	}()
}

func (a *Auditor) createSpecialTransaction(block *atypes.DistilledBlock) *atypes.SpecialTransaction {
	tx := &atypes.SpecialTransaction{
		ChainID: block.ChainID,
		Number:  block.Number,
		Hash:    block.Hash,
		Member:  a.AuditorID,
	}
	return tx
}

func (a *Auditor) interpret(block *atypes.DistilledBlock) *core.InterpretResult {

	// Prefix agreement executed on every block
	paResult := a.pa.DoPrefixAgreement(block)

	// Below are deprecated, because metric analysis is changed to behavior analysis and it has a separate implementation
	// Also, generator is moved to client-side
	// while, a metric analysis executed on a pre-defined time window
	// maResult := a.ma.DoMetricAnalysis(block)

	// tx := a.createSpecialTransaction(block)

	result := &core.InterpretResult{
		PAResult: paResult,
		// MAResult: maResult,
	}

	return result
}

func (a *Auditor) SentBefore(blockNum uint64) bool {
	if _, ok := a.SentBlocks[blockNum]; ok {
		return true
	}
	return false
}

func (a *Auditor) InterpretedBefore(blockNum uint64) bool {
	if _, ok := a.InterpretedBlocks[blockNum]; ok {
		return true
	}
	return false
}

// func (a *Auditor) GetAuditorSpec() {
//     a.logger.Debug("")
// }

// Decrecated: Replaced by Auditor.Run()
// AuditCore
// func StartAuditService(chainID, srcAddr, dstAddr string, provider adapter.BaasAdapter) {
//     NewBlockListener := provider.GetNewBlockListener(chainID, srcAddr)

//     var cs core.ConsensusSupport
//     numberOfPeers, _ := strconv.ParseUint(os.Getenv("CUSTOM_NUM_PEERS"), 10, 64)
//     quorumSize := util.GetQuorumSizeFromTotal(numberOfPeers)

//     var consensusTable core.ConsensusTable
//     consensusTable.PerMemberHighestVotedBlockTable = make(map[core.VoterType]core.HeightType)
//     var defaultSize int = 100000
//     consensusTable.AggregatedVoteTable = make([]uint64, defaultSize)
//     consensusTable.CommitBaseTable = make(map[core.HeightType]uint64)

//     pac := core.PrefixAgreement{
//         ChainID:                        chainID,
//         NumberOfPeers:                  numberOfPeers,
//         QuorumSize:                     quorumSize,
//         LatestCommittedHeight:          0,
//         LatestPreparedHeight:           0,
//         PreviousCommittedHeight:        0,
//         PreviousPreparedHeight:         0,
//         PreviousHeightWithSelfIssuedTx: 0,
//         ConsensusTable:                 consensusTable,
//         OldHeight:                      0,
//         NewHeight:                      0,
//     }

//     MemberID, _ := strconv.ParseUint(os.Getenv("MEMBERID"), 10, 64)
//     identity := core.MemberIdentity{
//         MemberName: os.Getenv("MEMBERNAME"),
//         MemberID:   MemberID,
//     }
//     mac := core.MetricAnalysis{}

//     cs.Initialize(chainID, identity, pac, mac)

//     // Core logics. It should be simple !
//     for {

//         block := <-NewBlockListener

//         // Interpret(block) or Evaluate(block)
//         // -- DoPrefixAgreement()
//         // -- DoMetricAnalysis()
//         SpecialTx := cs.Interpret(block)

//         // txid := provider.Submit(SpecialTx)
//         provider.Submit(SpecialTx)
//         // logger
//     }

//     // logger.Info("End of audit execution")
// }

func GetRedisDBClient(fetch_address string) (*redis.Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     fetch_address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := r.Ping().Result()
	if err != nil {
		logger.Error(err, pong)
	}

	return r, err
}
