package web

import (
	"time"

	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"
)

type HeartbeatRequest struct {
	SentAt time.Time `json:"sent_at"`
}

type UploadStatRequest struct {
	SentAt     time.Time `json:"sent_at"`
	UploadTime int64     `json:"upload_time"`
}

func ConvertToUploadStat(uploadStat UploadStatRequest) device.UploadStat {
	return device.UploadStat{
		ID:         "",
		SentAt:     uploadStat.SentAt,
		UploadTime: uploadStat.UploadTime,
	}
}

func ConvertToHeartbeat(heartbeat HeartbeatRequest) device.Heartbeat {
	return device.Heartbeat{
		ID:     "",
		SentAt: heartbeat.SentAt,
	}
}
