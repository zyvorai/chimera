package faults

import (
	"sync"
	"time"
)

type State struct {
	mu sync.RWMutex

	Latency        time.Duration `json:"-"`
	LatencyMS      int           `json:"latency_ms"`
	FailNext       int           `json:"fail_next"`
	FailStatus     int           `json:"fail_status"`
	NFCFailNext    int           `json:"nfc_fail_next"`
	NFCDropNext    int           `json:"nfc_drop_next"`
	NFCDropAfter   int64         `json:"nfc_drop_after_bytes"`
	BandwidthBPS   int64         `json:"bandwidth_bytes_per_sec"`
	Requests       uint64        `json:"requests"`
	InjectedFaults uint64        `json:"injected_faults"`
}

type Snapshot struct {
	LatencyMS      int    `json:"latency_ms"`
	FailNext       int    `json:"fail_next"`
	FailStatus     int    `json:"fail_status"`
	NFCFailNext    int    `json:"nfc_fail_next"`
	NFCDropNext    int    `json:"nfc_drop_next"`
	NFCDropAfter   int64  `json:"nfc_drop_after_bytes"`
	BandwidthBPS   int64  `json:"bandwidth_bytes_per_sec"`
	Requests       uint64 `json:"requests"`
	InjectedFaults uint64 `json:"injected_faults"`
}

func New() *State { return &State{FailStatus: 503} }

func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{
		LatencyMS:      s.LatencyMS,
		FailNext:       s.FailNext,
		FailStatus:     s.FailStatus,
		NFCFailNext:    s.NFCFailNext,
		NFCDropNext:    s.NFCDropNext,
		NFCDropAfter:   s.NFCDropAfter,
		BandwidthBPS:   s.BandwidthBPS,
		Requests:       s.Requests,
		InjectedFaults: s.InjectedFaults,
	}
}

func (s *State) Apply(in Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LatencyMS = in.LatencyMS
	s.Latency = time.Duration(in.LatencyMS) * time.Millisecond
	s.FailNext = in.FailNext
	s.FailStatus = in.FailStatus
	if s.FailStatus == 0 {
		s.FailStatus = 503
	}
	s.NFCFailNext = in.NFCFailNext
	s.NFCDropNext = in.NFCDropNext
	s.NFCDropAfter = in.NFCDropAfter
	s.BandwidthBPS = in.BandwidthBPS
}

func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Latency, s.LatencyMS = 0, 0
	s.FailNext, s.NFCFailNext, s.NFCDropNext = 0, 0, 0
	s.NFCDropAfter, s.BandwidthBPS = 0, 0
	s.FailStatus = 503
	s.Requests, s.InjectedFaults = 0, 0
}

// Before evaluates the current fault policy. drop is true only for requests
// selected by nfc_drop_next; this makes retry/resume scenarios deterministic.
func (s *State) Before(isNFC bool) (delay time.Duration, fail bool, status int, drop bool, dropAfter, bps int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Requests++
	delay = s.Latency
	status = s.FailStatus
	if status == 0 {
		status = 503
	}
	if isNFC && s.NFCFailNext > 0 {
		s.NFCFailNext--
		s.InjectedFaults++
		return delay, true, status, false, s.NFCDropAfter, s.BandwidthBPS
	}
	if s.FailNext > 0 {
		s.FailNext--
		s.InjectedFaults++
		return delay, true, status, false, s.NFCDropAfter, s.BandwidthBPS
	}
	if isNFC && s.NFCDropNext > 0 && s.NFCDropAfter > 0 {
		s.NFCDropNext--
		s.InjectedFaults++
		drop = true
	}
	return delay, false, status, drop, s.NFCDropAfter, s.BandwidthBPS
}
