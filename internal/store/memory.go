package store

import (
	"fmt"
	"sync"
	"time"

	devicePkg "github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"
)

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

func (s *MemoryStore) AddHeartbeat(deviceID string, timestamp time.Time) error {
	device, exists := s.GetDeviceState(deviceID)
	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	if device.HeartbeatsCount == 0 {
		device.FirstHeartbeat = timestamp
		device.LastHeartbeat = timestamp
	} else {
		if timestamp.After(device.LastHeartbeat) {
			device.LastHeartbeat = timestamp
		}
		if timestamp.Before(device.FirstHeartbeat) {
			device.FirstHeartbeat = timestamp
		}
	}
	device.HeartbeatsCount++
	return nil
}

func (s *MemoryStore) AddUploadTime(deviceID string, uploadTime int64) error {
	device, exists := s.GetDeviceState(deviceID)

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	device.UploadTimesSum += uploadTime
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
		return devicePkg.Stats{}, fmt.Errorf("device %s not found", deviceId)
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	// uptime = (sumHeartbeats / numMinutesBetweenFirstAndLastHeartbeat) * 100
	if state.HeartbeatsCount == 0 {
		return devicePkg.Stats{Uptime: 0, AvgUploadTime: "0"}, nil
	}

	// Minutes+1 because the window is inclusive of both first and last minute
	uptime := float64(state.HeartbeatsCount) / float64(state.LastHeartbeat.Sub(state.FirstHeartbeat).Minutes()+1) * 100

	// timeDuration = avg(arrayOfUploadTimeDurations)
	var timeDuration string
	if state.UploadTimesCount > 0 {
		timeDuration = time.Duration(float64(state.UploadTimesSum) / float64(state.UploadTimesCount)).String()
	}

	return devicePkg.Stats{Uptime: uptime, AvgUploadTime: timeDuration}, nil
}
