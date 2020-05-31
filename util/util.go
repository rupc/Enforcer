package util

import (
	"io/ioutil"
	"log"
	"math"
	"time"

	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
)

func CreateTmpFile(data, dir, nameFormat string) string {
	// Create tmp file where store metric results are into
	file, err := ioutil.TempFile(dir, nameFormat)
	if err != nil {
		log.Fatal(err)
	}

	if data != "" {
		file.WriteString(string(data))
	}
	// defer os.Remove(file.Name())

	// fmt.Println(file.Name()) // For example "dir/prefix054003078"
	return file.Name()
}
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func GetCommitteeQuorumSizeFromRatioAndTotal(committeeRatio float64, total uint64) uint64 {
	n := uint64(math.Floor(float64(total) * committeeRatio))
	return GetByzantineQuorumSizeFromTotal(n)
	// return
}

// GetQuorumSize calculates quorum size given n, total # of nodes
func GetByzantineQuorumSizeFromTotal(n uint64) uint64 {
	// return uint64(math.Floor(float64((2*(n-1))/3) + 1)) // q = 2f+1 = 2/3 * (n-1) + 1, (f=(n-1)/3)
	return 2*uint64(math.Floor(float64(n-1)/3)) + 1 // q = 2f+1 = 2/3 * (n-1) + 1, (f=(n-1)/3)
}

func IsInSlice(elem uint64, slice []uint64) bool {
	for _, v := range slice {
		if elem == v {
			return true
		}
	}
	return false
}

func newTransactionID(n uint64) string {
	return string(n) + "abcde"
}

func getTimestamp() *google_protobuf.Timestamp {
	var timestamp *google_protobuf.Timestamp
	return timestamp
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

// // Expected path="/etc/hyperledger/fabric/audit_peerlist
// // Return a table with pairs of (int id, name)
// func ReadPeerList(path string, num_peers uint64) map[uint64]string {
//     f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
//     if err != nil {
//         log.Fatalf("open file error: %v", err)
//         return nil
//     }
//     defer f.Close()

//     rd := bufio.NewReader(f)
//     var PeerIdNameTable_ = map[uint64]string{}
//     var i uint64
//     for i = 0; i < num_peers; i++ {
//         line, err := rd.ReadString('\n')
//         if err != nil {
//             // num_peers exceeds maximum number of lines
//             if err == io.EOF {
//                 break
//             }

//             log.Fatalf("read file line error: %v", err)
//             return nil
//         }
//         // _ = line // GET the line string
//         idname := strings.Split(line, " ")
//         id, _ := strconv.Atoi(idname[0])
//         name := strings.TrimSuffix(idname[1], "\n")
//         PeerIdNameTable_[uint64(id)] = name
//     }
//     return PeerIdNameTable_
// }

// func InitGlobalVariables() {
// GlobalPeerName = GetPeerEndpoinName()
// // GlobalBSPID = viper.GetString("audit.bsp.id")
// // GlobalBSPAddress = viper.GetString("audit.bsp.address")
// GlobalBSPID = viper.GetString("BSP_ID")
// GlobalBSPAddress = viper.GetString("audit.bsp.address")
// GlobalNumberOfPeers, _ = strconv.Atoi(os.Getenv("CUSTOM_NUM_PEERS"))
// num_peers, _ := strconv.Atoi(os.Getenv("CUSTOM_NUM_PEERS"))
// // committeeSize, _ := strconv.Atoi(os.Getenv("CORE_PEER_COMMITTEE_SIZE"))

// peerfilepath := viper.GetString("audit.peer_filepath")

// // NOTE format of peer ID. Is it peer0 or peer0.org1.example.com ?
// // Init PeerIdNameTable and PeerNameIdTable
// ConstructPeerIdNameTable(peerfilepath, uint64(num_peers))

// // Init globally accessible identifier for this peer
// GlobalPeerID = PeerNameIdTable[GlobalPeerName]
// // AuditQuorumSize = uint64(viper.GetInt("audit.audit_quorum_size"))

// // For use outside of audit module
// evalhelper.SetPeerName(GlobalPeerName)
// evalhelper.SetPeerId(GlobalPeerID)

// logger.Infof("Found peer list file path: %s, num_peers: %d, audit quorum: %d, peerName: %s, peerID: %s, bspid: %d, bspaddr: %s\n", peerfilepath, num_peers, AuditQuorumSize, GlobalPeerName, GlobalPeerID, GlobalBSPID, GlobalBSPAddress)
// }

// func GetPeerEndpoinName() string {
//     peerEndpoint, _ := peer.GetPeerEndpoint()
//     return peerEndpoint.Id.Name
// }

// func printBlockchainInfo() {
//     ch := peer.GetChannelsInfo()
//     for _, cid := range ch {
//         mspids := peer.GetMSPIDs(cid.ChannelId)
//         logger.Debugf("ChannelId %s, MSPIDs %v", cid.ChannelId, mspids)
//     }
//     localIP := peer.GetLocalIP()

//     // dsrl := mgmt.GetIdentityDeserializer("mychannel")
//     id, _ := mgmt.GetLocalMSP().GetIdentifier()
//     // sid, _ := mgmt.GetLocalMSP().GetSigningIdentity(id)

//     srid, _ := mgmt.GetLocalSigningIdentityOrPanic().Serialize()
//     // fmt.Printf("+%v\n+%v", id, sid)
//     // GetIdentifer() returns Org1MSP
//     // fmt.Printf("GetIdentifier(): +%v\n", id)
//     x := fmt.Sprintf("SerializedIdentity() %x\n", srid)
//     logger.Debug(x)
//     logger.Debug("size of iden", len(srid))
//     logger.Debug("size of iden(s)", len(x))

//     logger.Debugf("GetIdentifier(): %v\n", id)
//     logger.Debugf("LocalIP %v", localIP)
//     // panic("rest...")
// }

// func GetBlockHashByNumber(targetLedgerID string, height uint64) ([]byte, error) {
//     logger.Debugf("Access block[%d]", height)
//     ledger := peer.GetLedger(targetLedgerID)
//     block, err := ledger.GetBlockByNumber(height - 1)
//     if err != nil {
//         logger.Error(err)
//         return nil, err
//     }
//     return block.Header.Hash(), err
// }
