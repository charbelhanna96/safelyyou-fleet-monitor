package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/device"
	"github.com/charbelhanna96/safelyyou-fleet-monitor/internal/store"
)

var baseTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func minute(n int) time.Time {
	return baseTime.Add(time.Duration(n) * time.Minute)
}
func TestGetStats_UnknownDevice(t *testing.T) {
	s := store.NewMemoryStore()
	_, err := s.GetStats("ghost")
	if !errors.Is(err, store.ErrDeviceNotFound) {
		t.Errorf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestGetStats_ZeroHeartbeats(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	stats, err := s.GetStats("dev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Uptime != 0 {
		t.Errorf("expected uptime 0, got %f", stats.Uptime)
	}
}

func TestGetStats_OneHeartbeat_100Percent(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(0)})
	stats, err := s.GetStats("dev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Uptime != 100 {
		t.Errorf("expected 100, got %f", stats.Uptime)
	}
}

func TestGetStats_PerfectUptime(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	for i := 0; i < 5; i++ {
		s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(i)})
	}
	stats, _ := s.GetStats("dev-1")
	if stats.Uptime != 100 {
		t.Errorf("expected 100, got %f", stats.Uptime)
	}
}

func TestGetStats_MissingMinuteLowersUptime(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	// heartbeats at 0,1,3,4, missing minute 2, window is 5 slots
	for _, m := range []int{0, 1, 3, 4} {
		s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(m)})
	}
	stats, _ := s.GetStats("dev-1")
	expected := float64(4) / float64(5) * 100 // 80%
	if stats.Uptime != expected {
		t.Errorf("expected %f, got %f", expected, stats.Uptime)
	}
}

func TestGetStats_OutOfOrderHeartbeats(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	// send out of order, should still compute correct window
	for _, m := range []int{4, 1, 0, 3, 2} {
		s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(m)})
	}
	stats, _ := s.GetStats("dev-1")
	if stats.Uptime != 100 {
		t.Errorf("expected 100, got %f", stats.Uptime)
	}
}

func TestGetStats_DuplicateHeartbeatCapsAt100(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	// two heartbeats in minute 0, one in minute 1, window 2, count 3
	s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(0)})
	s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(0)})
	s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(1)})
	stats, _ := s.GetStats("dev-1")
	if stats.Uptime != 100 {
		t.Errorf("expected uptime capped at 100, got %f", stats.Uptime)
	}
}

func TestGetStats_AvgUploadTime(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	for _, ms := range []int64{100, 200, 300} {
		err := s.AddUploadTime(device.UploadStat{ID: "dev-1", UploadTime: ms})
		if err != nil {
			t.Fatalf("AddUploadTime error: %v", err)
		}
	}
	stats, _ := s.GetStats("dev-1")
	expected := float64(200)
	if stats.AvgUploadTime != expected {
		t.Errorf("expected avg %f, got %f", expected, stats.AvgUploadTime)
	}
}

func TestGetStats_ZeroUploadStats(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})
	stats, _ := s.GetStats("dev-1")
	if stats.AvgUploadTime != 0 {
		t.Errorf("expected 0, got %f", stats.AvgUploadTime)
	}
}

func TestGetStats_ConcurrentWrites(t *testing.T) {
	s := store.NewMemoryStore()
	s.AddDevices([]string{"dev-1"})

	done := make(chan struct{}, 200)
	for i := 0; i < 100; i++ {
		go func(i int) {
			s.AddHeartbeat(device.Heartbeat{ID: "dev-1", SentAt: minute(i)})
			done <- struct{}{}
		}(i)
		go func(i int) {
			s.AddUploadTime(device.UploadStat{ID: "dev-1", UploadTime: int64(i * 100)})
			done <- struct{}{}
		}(i)
	}
	for range 200 {
		<-done
	}
	// just verify no panic and valid range
	stats, err := s.GetStats("dev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Uptime < 0 || stats.Uptime > 100 {
		t.Errorf("uptime out of range: %f", stats.Uptime)
	}
}
