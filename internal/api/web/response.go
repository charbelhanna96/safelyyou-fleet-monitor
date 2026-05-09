package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"
)

type StatsResponse struct {
	Uptime        float64 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}

func ConvertToStatsResponse(stats device.Stats) StatsResponse {
	return StatsResponse{
		Uptime:        stats.Uptime,
		AvgUploadTime: time.Duration(stats.AvgUploadTime).String(),
	}
}
func WriteError(rw http.ResponseWriter, msg string, status int) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(map[string]string{"msg": msg})
}
