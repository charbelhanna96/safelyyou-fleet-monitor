package store

import (
	"fmt"
	"sync"
	"time"
)

type MemoryStore struct {
	mu      sync.RWMutex
	devices map[string]*DeviceState
}

type DeviceState struct {
	mu          sync.Mutex
	Heartbeats  []time.Time
	UploadTimes []int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		devices: make(map[string]*DeviceState),
	}
}

func (s *MemoryStore) AddDevice(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[deviceID]; !exists {
		s.devices[deviceID] = &DeviceState{}
	}
}

func (s *MemoryStore) AddHeartbeat(deviceID string, timestamp time.Time) error {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	device.Heartbeats = append(device.Heartbeats, timestamp)
	return nil
}

func (s *MemoryStore) AddUploadTime(deviceID string, uploadTime int64) error {
	s.mu.RLock()
	device, exists := s.devices[deviceID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	device.mu.Lock()
	defer device.mu.Unlock()
	device.UploadTimes = append(device.UploadTimes, uploadTime)
	return nil
}

func (s *MemoryStore) GetDeviceState(deviceID string) (*DeviceState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.devices[deviceID]
	return state, exists
}
