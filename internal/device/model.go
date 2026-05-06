package device

import "time"

type Heartbeat struct {
	ID     string
	SentAt time.Time
}

type Stats struct {
	Uptime        float64
	AvgUploadTime string
}

type UploadStat struct {
	ID         string
	SentAt     time.Time
	UploadTime int64
}
