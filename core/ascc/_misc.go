package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func Test_AuditTxes(args []string, stub shim.ChaincodeStubInterface) pb.Response {
	// args[1] assumes that an array of AuditTxes constructed in javascript lang.
	JSONStrings := args[1]

	var txes AuditTxes
	err := json.Unmarshal([]byte(JSONStrings), &txes)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}

	logger.Infof("Num of AuditTxes: %d", len(txes))
	logger.Infof("AuditTxes: %+v", txes)

	return shim.Success(nil)
}

// 2019-09-09 06:57:47.532 UTC [ascc] Invoke -> DEBU 095 function invoke args [test]
// 2019-09-09 06:57:47.532 UTC [ascc] Invoke -> INFO 096 test called !
// 2019-09-09 06:57:47.533 UTC [ascc] Test_QueryClientByTargetBlockNumber -> DEBU 097 zzz
// 2019-09-09 06:57:47.534 UTC [ascc] Test_QueryClientByTargetBlockNumber -> INFO 098 member [member client1 block 1] 4 [49]
// 2019-09-09 06:57:47.534 UTC [ascc] Test_QueryClientByTargetBlockNumber -> INFO 099 member [member client1 block 2] 4 [49]
// 2019-09-09 06:57:47.534 UTC [ascc] Test_QueryClientByTargetBlockNumber -> INFO 09a member [member client1 block 3] 4 [49]
// 2019-09-09 06:57:47.534 UTC [ascc] Test_QueryClientByTargetBlockNumber -> INFO 09b member [member client1 block 4] 4 [49]

func Test_QueryClientByTargetBlockNumber(stub shim.ChaincodeStubInterface) pb.Response {
	sampleKey := "zyzzyva"
	err := stub.PutState(sampleKey, []byte{'z', 'z', 'z'})
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	b, err := stub.GetState(sampleKey)
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	logger.Debugf("%s", string(b))

	// Create workload
	// 한 클라이언트가 투표한 블록 번호들 리턴, (client-centric)
	// objectType is a table name
	objectType := "client_voted_blocks_table"
	attributes1 := []string{"member", "client1", "block", "1"}
	attributes2 := []string{"member", "client1", "block", "2"}
	attributes3 := []string{"member", "client1", "block", "3"}
	attributes4 := []string{"member", "client1", "block", "4"}

	attributes5 := []string{"member", "client2", "block", "1"}
	attributes6 := []string{"member", "client2", "block", "2"}
	attributes7 := []string{"member", "client2", "block", "3"}
	attributes8 := []string{"member", "client2", "block", "4"}

	attributes9 := []string{"member", "client3", "block", "1"}
	attributes10 := []string{"member", "client3", "block", "2"}
	attributes11 := []string{"member", "client3", "block", "3"}
	attributes12 := []string{"member", "client3", "block", "4"}

	attributes13 := []string{"member", "client4", "block", "1"}
	attributes14 := []string{"member", "client4", "block", "2"}
	attributes15 := []string{"member", "client4", "block", "3"}
	attributes16 := []string{"member", "client4", "block", "4"}

	clientCompositeKey1, _ := stub.CreateCompositeKey(objectType, attributes1)
	clientCompositeKey2, _ := stub.CreateCompositeKey(objectType, attributes2)
	clientCompositeKey3, _ := stub.CreateCompositeKey(objectType, attributes3)
	clientCompositeKey4, _ := stub.CreateCompositeKey(objectType, attributes4)

	clientCompositeKey5, _ := stub.CreateCompositeKey(objectType, attributes5)
	clientCompositeKey6, _ := stub.CreateCompositeKey(objectType, attributes6)
	clientCompositeKey7, _ := stub.CreateCompositeKey(objectType, attributes7)
	clientCompositeKey8, _ := stub.CreateCompositeKey(objectType, attributes8)

	clientCompositeKey9, _ := stub.CreateCompositeKey(objectType, attributes9)
	clientCompositeKey10, _ := stub.CreateCompositeKey(objectType, attributes10)
	clientCompositeKey11, _ := stub.CreateCompositeKey(objectType, attributes11)
	clientCompositeKey12, _ := stub.CreateCompositeKey(objectType, attributes12)

	clientCompositeKey13, _ := stub.CreateCompositeKey(objectType, attributes13)
	clientCompositeKey14, _ := stub.CreateCompositeKey(objectType, attributes14)
	clientCompositeKey15, _ := stub.CreateCompositeKey(objectType, attributes15)
	clientCompositeKey16, _ := stub.CreateCompositeKey(objectType, attributes16)

	stub.PutState(clientCompositeKey1, []byte{'1'})
	stub.PutState(clientCompositeKey2, []byte{'1'})
	stub.PutState(clientCompositeKey3, []byte{'1'})
	stub.PutState(clientCompositeKey4, []byte{'1'})

	stub.PutState(clientCompositeKey5, []byte{'1'})
	stub.PutState(clientCompositeKey6, []byte{'1'})
	stub.PutState(clientCompositeKey7, []byte{'1'})
	stub.PutState(clientCompositeKey8, []byte{'1'})

	stub.PutState(clientCompositeKey9, []byte{'1'})
	stub.PutState(clientCompositeKey10, []byte{'1'})
	stub.PutState(clientCompositeKey11, []byte{'1'})
	stub.PutState(clientCompositeKey12, []byte{'1'})

	stub.PutState(clientCompositeKey13, []byte{'1'})
	stub.PutState(clientCompositeKey14, []byte{'1'})
	stub.PutState(clientCompositeKey15, []byte{'1'})
	stub.PutState(clientCompositeKey16, []byte{'1'})

	// client 1 필터링
	logger.Info("Filter blocks voted by client1")
	client1Iter, err := stub.GetStateByPartialCompositeKey(objectType, []string{"member", "client1", "block"})
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	defer client1Iter.Close()

	for client1Iter.HasNext() {
		KV, err := client1Iter.Next()
		t, attrs, err := stub.SplitCompositeKey(KV.Key)
		if err != nil {
			logger.Error(err)
			continue
		}
		logger.Info(t, attrs, len(attrs), string(KV.Value))
	}

	// client 2 필터링
	logger.Info("Filter blocks voted by client2")
	client2Iter, err := stub.GetStateByPartialCompositeKey(objectType, []string{"member", "client2", "block"})
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	defer client2Iter.Close()

	for client2Iter.HasNext() {
		KV, err := client2Iter.Next()
		t, attrs, err := stub.SplitCompositeKey(KV.Key)
		if err != nil {
			logger.Error(err)
			continue
		}
		logger.Info(t, attrs, len(attrs), string(KV.Value))
	}

	// assumes
	swap_to_block_prefix := func(attrs []string) []string {
		attrs[0], attrs[1], attrs[2], attrs[3] = attrs[2], attrs[3], attrs[0], attrs[1]
		return attrs
	}

	// block based filtering
	objectType = "block_votedby_client"
	blockCompositeKey1, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes1))
	blockCompositeKey2, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes2))
	blockCompositeKey3, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes3))
	blockCompositeKey4, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes4))

	blockCompositeKey5, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes5))
	blockCompositeKey6, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes6))
	blockCompositeKey7, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes7))
	blockCompositeKey8, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes8))

	blockCompositeKey9, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes9))
	blockCompositeKey10, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes10))
	blockCompositeKey11, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes11))
	blockCompositeKey12, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes12))

	blockCompositeKey13, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes13))
	blockCompositeKey14, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes14))
	blockCompositeKey15, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes15))
	blockCompositeKey16, _ := stub.CreateCompositeKey(objectType, swap_to_block_prefix(attributes16))

	stub.PutState(blockCompositeKey1, []byte{'1'})
	stub.PutState(blockCompositeKey2, []byte{'1'})
	stub.PutState(blockCompositeKey3, []byte{'1'})
	stub.PutState(blockCompositeKey4, []byte{'1'})

	stub.PutState(blockCompositeKey5, []byte{'1'})
	stub.PutState(blockCompositeKey6, []byte{'1'})
	stub.PutState(blockCompositeKey7, []byte{'1'})
	stub.PutState(blockCompositeKey8, []byte{'1'})

	stub.PutState(blockCompositeKey9, []byte{'1'})
	stub.PutState(blockCompositeKey10, []byte{'1'})
	stub.PutState(blockCompositeKey11, []byte{'1'})
	stub.PutState(blockCompositeKey12, []byte{'1'})

	stub.PutState(blockCompositeKey13, []byte{'1'})
	stub.PutState(blockCompositeKey14, []byte{'1'})
	stub.PutState(blockCompositeKey15, []byte{'1'})
	stub.PutState(blockCompositeKey16, []byte{'1'})

	// block1 에 투표한 클라이언트 필터링
	logger.Info("Filter clients, which voted to block1")
	block1Iter, err := stub.GetStateByPartialCompositeKey(objectType, []string{"block", "1"})
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	defer block1Iter.Close()

	for block1Iter.HasNext() {
		KV, err := block1Iter.Next()
		t, attrs, err := stub.SplitCompositeKey(KV.Key)
		if err != nil {
			logger.Error(err)
			continue
		}
		logger.Info(t, attrs, len(attrs), string(KV.Value))
	}

	// block2 에 투표한 클라이언트 필터링
	logger.Info("Filter clients, which voted to block2")
	block2Iter, err := stub.GetStateByPartialCompositeKey(objectType, []string{"block", "2"})
	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	defer block2Iter.Close()

	for block2Iter.HasNext() {
		KV, err := block2Iter.Next()
		t, attrs, err := stub.SplitCompositeKey(KV.Key)
		if err != nil {
			logger.Error(err)
			continue
		}
		logger.Info(t, attrs, len(attrs), string(KV.Value))
	}

	return shim.Success(nil)
}
