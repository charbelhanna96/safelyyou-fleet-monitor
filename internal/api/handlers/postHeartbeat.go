package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/api/web"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/store"
)

func (h *Handler) PostHeartbeat(rw http.ResponseWriter, request *http.Request) {
	// parse request body to get device ID and timestamp
	var heartbeatReq web.HeartbeatRequest
	deviceID := request.PathValue("device_id")

	if err := json.NewDecoder(request.Body).Decode(&heartbeatReq); err != nil {
		slog.Error("decode error", "error", err, "device_id", deviceID)
		web.WriteError(rw, "invalid request body", http.StatusBadRequest)
		return
	}

	heartbeat := web.ConvertToHeartbeat(heartbeatReq)
	heartbeat.ID = deviceID

	// add heartbeat to store
	err := h.store.AddHeartbeat(heartbeat)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			web.WriteError(rw, err.Error(), http.StatusNotFound)
		} else {
			web.WriteError(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
