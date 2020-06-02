package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/rupc/Enforcer/core/audit"
	"github.com/rupc/Enforcer/core/pa"
)

type AuditHandler struct {
	logger *flogging.FabricLogger
}

func GetAuditHandler() *AuditHandler {
	return &AuditHandler{
		logger: flogging.MustGetLogger("codit"),
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Serves as a control message
type AuditServiceRequest struct {
	// init:auditor
	// init:pa
	// init:adapter
	// query
	Cmd string `json:"cmd"`

	// cmd: init
	// required to firstly init auditor
	CreatorID          string   `json:"creator_id"`
	AuditorID          string   `json:"auditor_id"`
	TargetChainID      string   `json:"target_chain_id"` // target chain
	RecordChainID      string   `json:"record_chain_id"` // chain that blocks come from, usually ChainID
	Fidelity           string   `json:"fidelity"`
	TxSendAddress      string   `json:"tx_send_address"`
	StateSendAddress   string   `json:"state_send_address"`
	FetchMethod        string   `json:"fetch_method"`
	MetricFetchAddress string   `json:"metric_fetch_address"`
	ReportAddress      string   `json:"report_address"`
	FetchAddress       string   `json:"fetch_address"` // Block fetch address, e.g. redisdb for 141.223.121.56:12000
	MetricNames        []string `json:"metric_names"`

	// adapter-specific
	// After design modification, role of BaaS Adapter outsourced to outer SDK server
	Platform          string `json:"platform"`
	BlockFetchAddress string `json:"block_fetch_address"`

	// pa init
	PAChainID  string   `json:"pa_chain_id"`
	MemberName string   `json:"member_name"` // pa-specific identity
	MemberID   string   `json:"member_id"`
	NumOfPeers uint64   `json:"num_of_peers"`
	QuorumSize uint64   `json:"quorum_size"`
	MemberList []string `json:"member_list"`

	// params for SendAnalysis request
	AnalysisType string `json:"analysis_type"`
	StartBlkNum  string `json:"start_blk_num"`
	EndBlkNum    string `json:"end_blk_num"`
	WindowSize   string `json:"window_size"`

	// var params = {
	//   startNum: v[0],
	//   endNum: v[1],
	//   windowSize: v[2],
	// }
	// ma init
	// method string
	// ....
	// Fairness
	// Performance anomaly: correlation analysis
	// SLA Violation, QoS

	// Etc
	Version string `json:"version"`
}

type InitCmd struct {
}

type AuditServiceResponse struct {
	Msg string `json:"msg"`
}

func (h *AuditHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	h.logger.Info("audit service requested")

	var asr AuditServiceRequest
	decoder := json.NewDecoder(req.Body)

	if err := decoder.Decode(&asr); err != nil {
		h.sendResponse(resp, http.StatusBadRequest, err)
		return
	}

	req.Body.Close()
	cmd := strings.ToLower(asr.Cmd)
	switch req.Method {
	case http.MethodPost:
		switch cmd {
		case "init":
			h.handleInit(resp, asr)
			h.logger.Infof("Handle init request %+v", asr)
		// case "start":
		//     h.handleStart(resp, asr)
		//     h.logger.Infof("Handle start request %+v", asr)
		case "start_analysis":
			h.handleStartAnalysis(resp, asr)
			h.logger.Infof("Handle start_analysis request %+v", asr)

		default:
			h.sendResponse(resp, http.StatusBadRequest, "Unknown commands")
		}

	case http.MethodGet:
		h.sendResponse(resp, http.StatusOK, &AuditServiceRequest{Cmd: "MethodGet"})
	case http.MethodPut:
	default:
		err := fmt.Errorf("invalid request method: %s", req.Method)
		h.sendResponse(resp, http.StatusBadRequest, err)
	}
}

func (h *AuditHandler) sendResponse(resp http.ResponseWriter, code int, payload interface{}) {
	encoder := json.NewEncoder(resp)
	if err, ok := payload.(error); ok {
		payload = &ErrorResponse{Error: err.Error()}
	}

	resp.WriteHeader(code)

	resp.Header().Set("Content-Type", "application/json")
	if err := encoder.Encode(payload); err != nil {
		h.logger.Errorw("failed to encode payload", "error", err)
	}
}

// handle a start_analysis request
func (h *AuditHandler) handleStartAnalysis(resp http.ResponseWriter, r AuditServiceRequest) {

	if r.AnalysisType == "" ||
		r.StartBlkNum == "" ||
		r.EndBlkNum == "" ||
		r.WindowSize == "" {
		h.logger.Error("Init failed due to null fields in start_analysis request ")
		h.sendResponse(resp, http.StatusBadRequest, "start_analysis failed due to null fields in start_analysis")
		return
	}

	// Run analysis
	// get auditor name from pool
	auditor := audit.GetAuditorByID(r.AuditorID)
	if auditor == nil {
		h.sendResponse(resp, http.StatusBadRequest, "auditor doesn't exist")
		return
	}

	// send start analysis request to a AnalysisRequestChannel of the auditor

	// wait for the result
	// send its result to the client
	h.sendResponse(resp, http.StatusOK, "Finished analysis!")
}

// func (h *AuditHandler) handleStart(resp http.ResponseWriter, auditRequest AuditServiceRequest) {

// }

func (h *AuditHandler) handleInit(resp http.ResponseWriter, r AuditServiceRequest) {

	// platform := auditRequest.Platform
	// version := auditRequest.Version
	// blkAddr := auditRequest.BlockFetchAddress
	// txAddr := auditRequest.TxSendAddress
	// Validate nil inputs
	if r.AuditorID == "" ||
		r.CreatorID == "" ||
		r.TargetChainID == "" ||
		r.RecordChainID == "" ||
		r.TxSendAddress == "" ||
		r.Fidelity == "" ||
		r.StateSendAddress == "" ||
		r.FetchMethod == "" ||
		r.MetricFetchAddress == "" ||
		r.ReportAddress == "" ||
		r.MetricNames == nil {

		// iterate to find a null value in the struct fields
		v := reflect.ValueOf(r)
		fields := reflect.TypeOf(r)

		values := make([]interface{}, v.NumField())

		for i := 0; i < v.NumField(); i++ {
			values[i] = v.Field(i).Interface()
			field := fields.Field(i)
			if values[i] == "" {
				h.logger.Error("Empty value: ", field.Name, values[i])
			}
		}

		fmt.Println(values)
		h.logger.Error("Init failed due to null fields in init request")
		h.sendResponse(resp, http.StatusBadRequest, "Init failed due to null fields in init request")
		return
	}

	// Early validity check
	if !(r.FetchMethod == "redis" || r.FetchMethod == "leveldb" || r.FetchMethod == "prom") {
		h.logger.Error("Init failed due to invalid fetch method: expected redis or leveldb or prom, but got", r.FetchMethod)
		h.sendResponse(resp, http.StatusBadRequest, "Init failed due to null fields in init request")
		return
	}

	paConfig := &pa.PrefixAgreementConfig{
		ChainID:         r.TargetChainID,
		NumberOfMembers: uint64(len(r.MemberList)),
		MemberList:      []string{"peer1", "peer2", "peer3", "peer4"},
	}

	_, err := audit.InitializeAuditor(r.AuditorID, r.CreatorID, r.TargetChainID, r.RecordChainID, r.TxSendAddress, r.Fidelity, r.StateSendAddress, r.FetchMethod, r.MetricFetchAddress, r.ReportAddress, r.FetchAddress, r.MetricNames, paConfig)
	h.logger.Infof("Initialized auditor[%s], creatorID: %s, target: %s, record: %s, tx_send_address: %s, state_send_address:%s, metric_fetch_address: %s, metric_names: %+v", r.AuditorID, r.CreatorID, r.TargetChainID, r.RecordChainID, r.TxSendAddress, r.StateSendAddress, r.MetricFetchAddress, r.MetricNames)

	if err != nil {
		h.logger.Error(err)
		h.sendResponse(resp, http.StatusBadRequest, err)
		return
	}

	msg := fmt.Sprintf("%s[%s] registered successfully", r.AuditorID, r.CreatorID)
	h.sendResponse(resp, http.StatusOK, &AuditServiceResponse{Msg: msg})
}

func (h *AuditHandler) handleStop(resp http.ResponseWriter, auditRequest AuditServiceRequest) {
}
