package core

import "github.com/rupc/audit/core/pa"

type InterpretResult struct {
	PAResult *pa.PrefixAgreementResult
	MAResult *MetricAnalysisResult
}

// type HeightStatus int32
// type HeightType uint64
// type VoterType uint64

// const (
//     HEIGHT_PREPARED  HeightStatus = 1
//     HEIGHT_COMMITTED HeightStatus = 2
// )

// type MemberIdentity struct {
//     MemberName string
//     MemberID   uint64
// }

// type ConsensusSupport struct {
//     ChainID   string
//     pac       PrefixAgreement
//     mac       MetricAnalysis
//     Identity  MemberIdentity
//     batchSize uint64
//     logger    *flogging.FabricLogger
// }

// // Initialize is chain-specific initialization
// func (cs *ConsensusSupport) Initialize(ChainID string, identity MemberIdentity, pac PrefixAgreement, mac MetricAnalysis) {
//     // initializeDefault()
//     cs.ChainID = ChainID
//     cs.logger = flogging.MustGetLogger(ChainID)

//     cs.pac = pac
//     cs.mac = mac
// }

// func (cs *ConsensusSupport) Interpret(block *DistilledBlock) SpecialTransaction {
//     cs.pac.DoPrefixAgreement(block)
//     cs.mac.DoMetricAnalysis(block)
//     tx := SpecialTransaction{
//         ChainID:  cs.ChainID,
//         Number:   block.Number,
//         Hash:     block.Hash,
//         Member:   cs.Identity.MemberName,
//     }
//     return tx
// }

// func GetPeerEndpoinName() string {
//     peerEndpoint, _ := peer.GetPeerEndpoint()
//     return peerEndpoint.Id.Name
// }
