package main

import (
	"sync"
	"time"
)

type TPSMonitor struct {
	mu              sync.RWMutex
	lastWindow      time.Time
	windowCommitted int64
	totalCommitted  int64
	totalLatencyNs  int64
	latencyCount    int64
}

func NewTPSMonitor() *TPSMonitor {
	now := time.Now()
	return &TPSMonitor{lastWindow: now}
}

func (m *TPSMonitor) RecordCommitted(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.windowCommitted += int64(n)
	m.totalCommitted += int64(n)
}

func (m *TPSMonitor) RecordConfirmLatency(d time.Duration) {
	if d < 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalLatencyNs += d.Nanoseconds()
	m.latencyCount++
}

func (m *TPSMonitor) Snapshot(shardID, nodeID int, pending int, latestHeight int64) STCMetricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(m.lastWindow).Seconds()
	tps := 0.0
	if elapsed > 0 {
		tps = float64(m.windowCommitted) / elapsed
	}
	avgLatencyMs := 0.0
	if m.latencyCount > 0 {
		avgLatencyMs = float64(m.totalLatencyNs/int64(time.Millisecond)) / float64(m.latencyCount)
	}
	m.windowCommitted = 0
	m.lastWindow = now
	return STCMetricsSnapshot{
		ShardID:             shardID,
		NodeID:              nodeID,
		TPS:                 tps,
		AvgConfirmLatencyMs: avgLatencyMs,
		TotalCommittedTx:    m.totalCommitted,
		PendingTx:           pending,
		LatestHeight:        latestHeight,
		LastUpdatedUnixNano: now.UnixNano(),
	}
}
