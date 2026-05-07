package handlers

import "github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"

type DataAccess interface {
	AddHeartbeat(hb device.Heartbeat) error
	AddUploadTime(stat device.UploadStat) error
	GetStats(deviceID string) (device.Stats, error)
}

type Handler struct {
	store DataAccess
}

func NewHandler(s DataAccess) *Handler {
	return &Handler{store: s}
}
