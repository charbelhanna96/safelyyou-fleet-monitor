package store

import (
	"errors"
	"sync"
	"time"

	devicePkg "github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"
)

var ErrDeviceNotFound = errors.New("device not found")

type MemoryStore struct {
	mu           sync.RWMutex
	deviceStates map[string]*DeviceState
}

type DeviceState struct {
	// RWMutex because we want to allow concurrent reads of the device state while still allowing updates to be made safely
	mu               sync.RWMutex
	FirstHeartbeat   time.Time
	LastHeartbeat    time.Time
	HeartbeatsCount  int
	UploadTimesSum   int64
	UploadTimesCount int
}

func NewMemoryStore() *MemoryStore {
	// using make to initialize the map to avoid nil map assignment errors when adding devices
	return &MemoryStore{
		deviceStates: make(map[string]*DeviceState),
	}
}

func (s *MemoryStore) AddDevice(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.deviceStates[deviceID]; !exists {
		s.deviceStates[deviceID] = &DeviceState{}
	}
}

func (s *MemoryStore) AddHeartbeat(hb devicePkg.Heartbeat) error {
	device, exists := s.GetDeviceState(hb.ID)
	if !exists {
		return ErrDeviceNotFound
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	if device.HeartbeatsCount == 0 {
		device.FirstHeartbeat = hb.SentAt
		device.LastHeartbeat = hb.SentAt
	} else {
		if hb.SentAt.After(device.LastHeartbeat) {
			device.LastHeartbeat = hb.SentAt
		}
		if hb.SentAt.Before(device.FirstHeartbeat) {
			device.FirstHeartbeat = hb.SentAt
		}
	}
	device.HeartbeatsCount++
	return nil
}

func (s *MemoryStore) AddUploadTime(ut devicePkg.UploadStat) error {
	device, exists := s.GetDeviceState(ut.ID)

	if !exists {
		return ErrDeviceNotFound
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	device.UploadTimesSum += ut.UploadTime
	device.UploadTimesCount++
	return nil
}

func (s *MemoryStore) GetDeviceState(deviceID string) (*DeviceState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.deviceStates[deviceID]
	return state, exists
}

func (s *MemoryStore) GetStats(deviceId string) (devicePkg.Stats, error) {
	state, exists := s.GetDeviceState(deviceId)
	if !exists {
		return devicePkg.Stats{}, ErrDeviceNotFound
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	// uptime = (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
	if state.HeartbeatsCount == 0 {
		return devicePkg.Stats{Uptime: 0, AvgUploadTime: 0}, nil
	}

	lastMinute := state.LastHeartbeat.Unix() / 60
	firstMinute := state.FirstHeartbeat.Unix() / 60
	numMinutes := (lastMinute - firstMinute) + 1

	// uptime is capped at 100% in case duplicate heartbeats are received (same minute heartbeats)
	uptime := (float64(state.HeartbeatsCount) / float64(numMinutes)) * 100
	if uptime > 100 {
		uptime = 100
	}

	// timeDuration = avg(arrayOfUploadTimeDurations)
	var timeDuration float64
	if state.UploadTimesCount > 0 {
		timeDuration = float64(state.UploadTimesSum) / float64(state.UploadTimesCount)
	}

	return devicePkg.Stats{Uptime: uptime, AvgUploadTime: timeDuration}, nil
}
