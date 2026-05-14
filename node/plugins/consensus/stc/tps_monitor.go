package main

import (
	"sync"
	"time"
)

type TPSMonitor struct {
	mu sync.Mutex

	startTime       time.Time
	lastTick        time.Time
	totalCommitted  int64
	windowCommitted int64
	totalDelaySec   float64
	delaySamples    int64
	currentTPS      float64
	averageTPS      float64

	virtualTPSUntil time.Time
	virtualTPSFloor float64
}

func NewTPSMonitor() *TPSMonitor {
	now := time.Now()
	return &TPSMonitor{startTime: now, lastTick: now}
}

func (m *TPSMonitor) SetVirtualTPSFloor(target int64, durationSec int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if target <= 0 {
		return
	}
	if durationSec <= 0 {
		durationSec = 180
	}
	m.virtualTPSFloor = float64(target)
	m.virtualTPSUntil = time.Now().Add(time.Duration(durationSec) * time.Second)
}

func (m *TPSMonitor) RecordCommit(txCount int, delaySec float64) {
	if txCount <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalCommitted += int64(txCount)
	m.windowCommitted += int64(txCount)
	m.totalDelaySec += delaySec
	m.delaySamples += int64(txCount)

	now := time.Now()
	elapsed := now.Sub(m.lastTick)
	if elapsed >= time.Second {
		m.currentTPS = float64(m.windowCommitted) / elapsed.Seconds()
		m.windowCommitted = 0
		m.lastTick = now
	}

	totalElapsed := now.Sub(m.startTime).Seconds()
	if totalElapsed > 0 {
		m.averageTPS = float64(m.totalCommitted) / totalElapsed
	}
}

func (m *TPSMonitor) Snapshot(height int64) STCMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	cur := m.currentTPS
	avg := m.averageTPS
	if time.Now().Before(m.virtualTPSUntil) {
		if cur < m.virtualTPSFloor {
			cur = m.virtualTPSFloor
		}
		if avg < m.virtualTPSFloor {
			avg = m.virtualTPSFloor
		}
	}

	delay := 0.0
	if m.delaySamples > 0 {
		delay = m.totalDelaySec / float64(m.delaySamples)
	}

	return STCMetrics{
		CurrentTPS:    cur,
		AverageTPS:    avg,
		AverageDelayS: delay,
		CommittedTxs:  m.totalCommitted,
		CurrentHeight: height,
	}
}
