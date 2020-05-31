package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

var logger *flogging.FabricLogger

const (
	// Including function argument
	storeAuditTxNumArg                     = 6
	QueryAuditTxByTargetBlockNumber_NumArg = 2
	QueryClientByTargetBlockNumber_NumArg  = 2
	QueryAuditTxByClient_NumArg            = 2
	QueryClientState_NumArg                = 2
	Query_NumArg                           = 2

	objectType_client_prefix = "client_voted_blocks_table"
	objectType_block_prefix  = "block_votedby_client"
)

type ascc struct{}

type AuditTx struct {
	TxID string `json:"tx_id"`
	// ChannelID string `json:"channel_id"`
	ChainID string `json:"chain_id"`
	Number  string `json:"number"`
	Hash    string `json:"hash"`

	// Member can be either clientID or peerID
	Member string `json:"member"`
	Nonce  string `json:"nonce"`
}

type Querier struct {
}

type AuditTxes []AuditTx

// New returns an implementation of the chaincode interface.
func New() shim.Chaincode {
	logger = flogging.MustGetLogger("ascc")
	logger.Info("ASCC constructed!")
	return &ascc{}
}

var initCnt int = 0
var once sync.Once

// Init implements the chaincode shim interface
func (s *ascc) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("ASCC start to initialize")

	// DEBUG Seems to be fabric-bug, by executing Init() twice
	once.Do(func() {
		t := time.Now()

		logger.Infof("[%d]ASCC initialized at %s", initCnt, t.Format(time.RFC3339))
		initCnt++
		// testAuditor()
	})

	return shim.Success(nil)
}

// func testAuditor() {
// paConfig := pa.PrefixAgreementConfig{
//     ChainID:         "mychannel",
//     NumberOfMembers: 4,
// }

// pa := pa.InitializePrefixAgreement(paConfig)

// r := server.AuditServiceRequest{
// AuditorID:          "auditor9999",
// CreatorID:          "client9999",
// TargetChainID:      "basschain",
// RecordChainID:      "",
// TxSendAddress:      "",
// StateSendAddress:   "",
// MetricFetchAddress: "",
// ReportAddress:      "",
// MetricNames:        nil,
// }

// _, err := audit.InitializeAuditor(r.AuditorID, r.CreatorID, r.TargetChainID, r.RecordChainID, r.TxSendAddress, r.StateSendAddress, r.MetricFetchAddress, r.ReportAddress, r.MetricNames)

// if err != nil {
// logger.Error(err)
// return
// }

// logger.Infof("Initialized auditor[%s], creatorID: %s, target: %s, record: %s, tx_send_address: %s, state_send_address:%s, metric_fetch_address: %s, metric_names: %+v", r.AuditorID, r.CreatorID, r.TargetChainID, r.RecordChainID, r.TxSendAddress, r.StateSendAddress, r.MetricFetchAddress, r.MetricNames)
// }

// var AuditTxAttributes = []string{"chainID", "number", "hash", "member"}
// Invoke implements the chaincode shim interface

// Invocation test succeeded!
// 2019-08-13 07:41:49.886 UTC [ascc] Invoke -> DEBU a029 ASCC invoked !
// 2019-08-13 07:41:49.886 UTC [ascc] Invoke -> DEBU a02a invoke [storeAuditTx mychannel 10 0xabcdef peer1 ScABId74pR]

func (s *ascc) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	function, args := stub.GetFunctionAndParameters()
	logger.Debug("function", function, "args", args)

	if len(args) == 0 {
		return shim.Error("At least 1 argument needed")
	}
	// storeAuditTx,
	// - QueryAuditTxByTargetBlockNumber(number),
	// - QueryAuditTxByClient(member, number)
	// - QueryClientByTargetBlockNumber(number)
	// - QueryClientState(member)

	logger.Info(args[0], "called !")

	switch args[0] {
	case "storeAuditTx":
		response = storeAuditTx(args, stub)
	// case "queryAllAuditTx":
	//     response = QueryAllAuditTx(args, stub)
	case "QueryAuditTxByTargetBlockNumber":
		response = QueryAuditTxByTargetBlockNumber(args, stub)
	case "QueryAuditTxByClient":
		response = QueryAuditTxByClient(args, stub)
	// case "QueryClientListByTargetBlockNumber":
	//     response = QueryClientListByTargetBlockNumber(args, stub)
	// case "QueryClientState":
	//     response = QueryClientState(args, stub)
	// case "test":
	//     response = Test_QueryClientByTargetBlockNumber(stub)
	// case "test_AuditTxes":
	// response = Test_AuditTxes(args, stub)

	default:
		logger.Error("No matching function")
		response = shim.Error("No matching function")
	}

	return response
}

// storeAuditTx stores auditTx
// args[0]: storeAuditTx
// args[1]: Target ChainID
// args[2]: Block nubmer of target chainID
// args[3]: hash of block number
// args[4]: member id (string)
// args[5]: nonce
func storeAuditTx(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) != storeAuditTxNumArg {
		errMsg := fmt.Sprintf("storeAuditTx requires %d arguments", storeAuditTxNumArg)
		logger.Error(errMsg)
		return shim.Error(errMsg)
	}

	var err error

	auditTx := AuditTx{
		TxID:    stub.GetTxID(),
		ChainID: args[1],
		Number:  args[2],
		Hash:    args[3],
		Member:  args[4],
		Nonce:   args[5],
	}

	// Store AuditTx as TxID as prefix
	err = StoreAuditTxByTxId(args, auditTx, stub)
	if err != nil {
		logger.Error(err.Error())
		return shim.Error(err.Error())
	}

	// Store AuditTx as block number as prefix
	err = StoreAuditTxByBlockNumber(args, auditTx, stub)
	if err != nil {
		logger.Error(err.Error())
		return shim.Error(err.Error())
	}

	// Store AuditTx as client ID as prefix
	err = StoreAuditTxByClient(args, auditTx, stub)
	if err != nil {
		logger.Error(err.Error())
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// 주어진 블록 번호에 투표한 클라이언트 집합 리턴 용도
func StoreAuditTxByBlockNumber(args []string, auditTx AuditTx, stub shim.ChaincodeStubInterface) error {

	// objectType: "client_by_targetblocknumber"
	// key: "block".number."member".member
	// value: ""
	objectType := "block_votedby_client"
	attributes := []string{"block", auditTx.Number, "member", auditTx.Member}
	key, err := stub.CreateCompositeKey(objectType, attributes)

	if err != nil {
		return err
	}

	err = stub.PutState(key, []byte(auditTx.TxID))
	logger.Debugf("%s DB[%s]=%v", args[0], key, auditTx)

	if err != nil {
		return err
	}

	return nil
}

// StoreClientState 는 주어진 클라이언트의 상태 정보(클라이언트가 과거에 투표한 가장 큰 블록 번호) 리턴
// objectType: "client_state"
// key: "member".member."highest_block".number
// value: ""
func StoreAuditTxByClient(args []string, auditTx AuditTx, stub shim.ChaincodeStubInterface) error {
	objectType := "client_voted_blocks"
	attributes := []string{"member", auditTx.Member, "block", auditTx.Number}
	key, err := stub.CreateCompositeKey(objectType, attributes)

	if err != nil {
		return err
	}

	err = stub.PutState(key, []byte(auditTx.TxID))

	if err != nil {
		return err
	}

	logger.Debugf("%s DB[%s]=%v", args[0], key, auditTx)

	return nil
}

// args[1]: number
func QueryAuditTxByTargetBlockNumber(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) != Query_NumArg {
		errMsg := fmt.Sprintf("%s requires %d arguments", args[0], Query_NumArg)
		logger.Error(errMsg)
		return shim.Error(errMsg)
	}

	objectType := objectType_block_prefix
	keys := []string{"block", args[1]}

	keysIter, err := stub.GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		logger.Error("Failed at getting keysIter")
	}
	defer keysIter.Close()

	// var keys []string

	// Result is an array of clinet IDs that had voted for the block(in arg[1])
	var client string
	var clients []string

	for keysIter.HasNext() {
		KV, err := keysIter.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		key := KV.Key
		t, attrs, err := stub.SplitCompositeKey(key)
		if err != nil {
			logger.Error(err)
			continue
		}
		value := KV.Value
		logger.Info(t, attrs, value)

		client = attrs[3]
		clients = append(clients, client)
	}

	clientBytes, err := json.Marshal(clients)

	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success(clientBytes)
}

// QueryAuditTxByClient 는 특정 클라이언트가 투표했던 블록들 리턴
// GetStateByPartialCompositeKey() 사용하는게 핵심
// args[1]: member
// Corresponds to StoreClient

func QueryAuditTxByClient(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	if len(args) != Query_NumArg {
		errMsg := fmt.Sprintf("%s requires %d arguments", args[0], Query_NumArg)
		logger.Error(errMsg)
		return shim.Error(errMsg)
	}

	objectType := objectType_client_prefix
	keys := []string{"member", args[1]}

	keysIter, err := stub.GetStateByPartialCompositeKey(objectType, keys)
	if err != nil {
		logger.Error("Failed at getting keysIter")
	}
	defer keysIter.Close()

	// var keys []string

	// Result is an array of clinet IDs that had voted for the block(in arg[1])
	var client string
	var clients []string

	for keysIter.HasNext() {
		KV, err := keysIter.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		key := KV.Key
		t, attrs, err := stub.SplitCompositeKey(key)
		if err != nil {
			logger.Error(err)
			continue
		}
		value := KV.Value
		logger.Info(t, attrs, value)

		client = attrs[3]
		clients = append(clients, client)
	}

	clientBytes, err := json.Marshal(clients)

	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	return shim.Success(clientBytes)
}

// args[1]: member
// func QueryClientState(args []string, stub shim.ChaincodeStubInterface) pb.Response {
// if len(args) != Query_NumArg {
//     errMsg := fmt.Sprintf("%s requires %d arguments", args[0], Query_NumArg)
//     logger.Error(errMsg)
//     return shim.Error(errMsg)
// }
// member := args[1]
// return shim.Success(nil)
// }

// func QueryAllAuditTx(args []string, stub shim.ChaincodeStubInterface) pb.Response {
//     keysIter, err := stub.GetStateByPartialCompositeKey("AuditTx", nil)
//     if err != nil {
//         logger.Error("Failed at getting keysIter")
//     }
//     defer keysIter.Close()

//     // var keys []string

//     var tx AuditTx
//     var txes AuditTxes

//     for keysIter.HasNext() {
//         KV, err := keysIter.Next()

//         if err != nil {
//             return shim.Error(err.Error())
//         }

//         // key := KV.Key
//         value := KV.Value

//         err = json.Unmarshal(value, &tx)

//         if err != nil {
//             logger.Error("AuditTx marshal failed: ", err)
//             continue
//         }

//         logger.Debugf("append %+v", tx)
//         txes = append(txes, tx)

//     }

//     auditJson, err := json.Marshal(txes)

//     if err != nil {
//         logger.Error("Failed at marshaling all AuditTxes")
//         return shim.Error("Failed at marshaling all AuditTxes")
//     }

//     return shim.Success(auditJson)

//     // buf := &bytes.Buffer{}
//     // gob.NewEncoder(buf).Encode(keys)
//     // bs := buf.Bytes()
//     // fmt.Printf("%q", bs)

//     // return shim.Success(values)
//     // // return shim.Success(bs)
//     // return shim.Success(nil)
// }

// StoreAuditTx 은 그대로 이어 붙여서 저장함
func StoreAuditTxByTxId(args []string, auditTx AuditTx, stub shim.ChaincodeStubInterface) error {
	args = GetFields(auditTx)
	key, err := stub.CreateCompositeKey("AuditTx", args)
	if err != nil {
		return err
	}

	b, err := json.Marshal(auditTx)

	if err != nil {
		return err
	}

	err = stub.PutState(key, b)
	logger.Debugf("%s DB[%s]=%v", args[0], key, auditTx)

	if err != nil {
		return err
	}

	return nil
}

func main() {
	logger.Debug("ASCC Main function called!")
}
