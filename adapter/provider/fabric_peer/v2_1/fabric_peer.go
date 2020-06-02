package fabric_peer

import (
	"encoding/json"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/rupc/Enforcer/atypes"
	"github.com/rupc/Enforcer/logging/flogging"
)

type FabricPeerAdapterV2_1 struct {
	ChainID string
}

// copied from auditcc.go for json parsing
type AuditTrail struct {
	ChainID string `json:"chain_id"`
	Number  string `json:"number"`
	Hash    string `json:"hash"`
	Member  string `json:"member"`
	Nonce   string `json:"nonce"`
}

var logger flogging.FabricLogger

func (f *FabricPeerAdapterV2_1) Extract(block *common.Block) ([]*atypes.SpecialTransaction, error) {
	var TxSet []*atypes.SpecialTransaction
	var err error

	for _, d := range block.Data.Data {
		if env, err := protoutil.GetEnvelopeFromBlock(d); err != nil {
			if err != nil {
				logger.Warningf("Could not unmarshal Envelope, err %s", err)
				continue
			}

		} else if env != nil {
			var payload *common.Payload

			payload, err := protoutil.UnmarshalPayload(env.Payload)
			if err != nil {
				logger.Errorf("GetPayload returns err %s", err)
				continue
			}

			tx, err := protoutil.UnmarshalTransaction(payload.Data)

			for _, act := range tx.Actions {
				ccActionPayload, err := protoutil.UnmarshalChaincodeActionPayload(act.Payload)
				// ccProposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(ccActionPayload.ChaincodeProposalPayload)

				if err != nil {
					logger.Warningf("Could not unmarshal chaincode action payload, err %s", err)
				}

				prp, err := protoutil.UnmarshalProposalResponsePayload(ccActionPayload.Action.ProposalResponsePayload)
				if err != nil {
					logger.Warningf("Could not unmarshal proposal response payload, err %s", err)
				}

				ccAction, err := protoutil.UnmarshalChaincodeAction(prp.Extension)
				if err != nil {
					logger.Warningf("Could not unmarshal chaincode header extension, err %s", err)
				}

				// Note that result fields corresponds to rwset.TxReadWriteSet (data_model, ns_rwset)
				txrw_set := &rwset.TxReadWriteSet{}
				proto.Unmarshal(ccAction.Results, txrw_set)

				// target block number, target member
				var number string
				var member string

				for _, rwset := range txrw_set.NsRwset {
					// Unmarshall rwset
					kv := &kvrwset.KVRWSet{}
					err := proto.Unmarshal(rwset.Rwset, kv)
					if err != nil {
						logger.Errorf("Could not unmarshal KVRWSet due to %s", err)
						continue
					}
					value := kv.Writes[0].Value
					auditcc := AuditTrail{}
					err = json.Unmarshal([]byte(value), &auditcc)
					if err != nil {
						logger.Errorf("Could not unmarshal KVRWSet due to %s", err)
						continue
					}

					member = auditcc.Member
					number = auditcc.Number
				}

				// Detect special transaction
				chdr, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
				channel := chdr.ChannelId
				if ccAction.ChaincodeId.Name == "auditcc" {
					logger.Infof("detect auditcc")
					number, err := strconv.ParseUint(number, 10, 64)
					if err != nil {
						logger.Errorf("Could not parse block number? WTF? %s", err)
					}
					tx := &atypes.SpecialTransaction{
						ChainID:   channel,
						Number:    number,
						Hash:      string(block.Header.DataHash),
						Member:    member,
						Signature: []byte{},
					}
					TxSet = append(TxSet, tx)
				}
			}

			if err != nil {
				logger.Warningf("Could not unmarshal channel header, err %s, skipping", err)
				continue
			}

		}
	}
	return TxSet, err
}

func (f *FabricPeerAdapterV2_1) Distill(block *common.Block) (*atypes.DistilledBlock, error) {
	txSet, err := f.Extract(block)
	if err != nil {
		logger.Errorf("Could not extract special transactions, err %s", err)
		return nil, err
	}

	distilledBlock := &atypes.DistilledBlock{
		Number:     block.Header.Number,
		ChainID:    f.ChainID,
		Hash:       string(block.Header.DataHash),
		TxSet:      txSet,
		TotalTxNum: len(block.Data.Data),
	}

	return distilledBlock, err
}

func IsSpecialTransaction() bool {
	return true
}

func ExtractMember() string {
	return ""
}

// def is_ascc(self, tx):
// if tx["payload"]["data"]["actions"][0]["payload"]["action"]["proposal_response_payload"]["extension"]["chaincode_id"]["name"] == "ascc":
// return True
// else:
// return False

// def extract_member_from_tx(self, tx):
// cc_value = tx["payload"]["data"]["actions"][0]["payload"]["action"]["proposal_response_payload"][
// "extension"]["results"]["ns_rwset"][0]["rwset"]["writes"][0]["value"]
// cc_value_json = json.loads(cc_value)
// return cc_value_json["member"]

// def extract_field_from_tx(self, tx):
// cc_value = tx["payload"]["data"]["actions"][0]["payload"]["action"]["proposal_response_payload"][
// "extension"]["results"]["ns_rwset"][0]["rwset"]["writes"][0]["value"]
// cc_value_json = json.loads(cc_value)
// return cc_value_json["member"], cc_value_json["number"]
