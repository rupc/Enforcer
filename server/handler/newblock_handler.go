package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rupc/audit/atypes"
	"github.com/rupc/audit/core/audit"
	"github.com/rupc/audit/logging/flogging"
)

type NewBlockHandler struct {
	logger *flogging.FabricLogger
}

func GetNewBlockHandler() *NewBlockHandler {
	return &NewBlockHandler{
		logger: flogging.MustGetLogger("newblock"),
	}

}

func (h *NewBlockHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// h.logger.Info("new block request")
	switch req.Method {
	case http.MethodPost:
		// h.logger.Info("post")
		// var dblock tesblock
		// var dblock atypes.DistilledBlockStub
		var dblock atypes.NewBlockMessage
		// dblock.TxSet = make([]*atypes.SpecialTransaction, 10)

		fmt.Printf("%+v\n", req.Body)
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&dblock); err != nil {
			h.sendResponse(resp, http.StatusBadRequest, "block marshal failure")
			h.logger.Error(err)
			return
		}
		req.Body.Close()

		// Input validation
		if dblock.AuditorID == "" {
			h.logger.Errorf("AuditorID[%s] is empty", dblock.AuditorID)
			h.sendResponse(resp, http.StatusBadRequest, "AuditorID is empty!")
			return
		}

		if dblock.Content.ChainID == "" {
			h.logger.Errorf("chainID[%s] is empty", dblock.Content.ChainID)
			h.sendResponse(resp, http.StatusBadRequest, "ChainID is empty!")
			return
		}

		if !audit.IsValidAuditor(dblock.AuditorID) {
			h.logger.Errorf("AuditorID[%s] is invalid", dblock.AuditorID)
			h.sendResponse(resp, http.StatusBadRequest, "Invalid auditor!")
			return
		}

		auditor := audit.GetAuditorByID(dblock.AuditorID)

		h.logger.Infof("Send a new block to auditor[%s]", auditor.AuditorID)
		// auditor.BlockArrivalChannel <- &dblock.Content
		auditor.NewBlockArrivalEvent(&dblock.Content)

		// Work with dblock
		// Send dblock via channel
		// chainID multiplexing

		// Send a new block to audit core via a channel
		//

		h.sendResponse(resp, http.StatusOK, "New Block will be audited")
		// echo request
		h.logger.Infof("Requested spec %+v", dblock)
		for i := 0; i < len(dblock.Content.TxSet); i++ {
			h.logger.Infof("SpecialTx: %+v", dblock.Content.TxSet[i])
		}
	case http.MethodPut:
		h.logger.Info("http.MethodPut")
	case http.MethodGet:
		h.logger.Info("http.MethodGet")
	default:
		h.sendResponse(resp, http.StatusBadRequest, "Unknown commands")
	}
}

func (h *NewBlockHandler) sendResponse(resp http.ResponseWriter, code int, payload interface{}) {
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
