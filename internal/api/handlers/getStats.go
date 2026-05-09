package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/api/web"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/store"
)

func (h *Handler) GetStats(rw http.ResponseWriter, request *http.Request) {
	deviceID := request.PathValue("device_id")

	stats, err := h.store.GetStats(deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			web.WriteError(rw, err.Error(), http.StatusNotFound)
		} else {
			web.WriteError(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(web.ConvertToStatsResponse(stats))
}
