package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rupc/Enforcer/core/audit"
	"github.com/rupc/Enforcer/logging/flogging"
)

type QueryHandler struct {
	logger *flogging.FabricLogger
}

func GetQueryHandler() *QueryHandler {
	return &QueryHandler{
		logger: flogging.MustGetLogger("query"),
	}
}
func (h *QueryHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		ids := audit.GetAuditorIDs()
		h.logger.Info("get", ids, len(ids))
		h.sendResponse(resp, http.StatusOK, ids)
	default:
		h.sendResponse(resp, http.StatusBadRequest, "Unknown commands")
	}
}
func (h *QueryHandler) sendResponse(resp http.ResponseWriter, code int, payload interface{}) {
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
