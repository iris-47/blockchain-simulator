package main

import (
	"fmt"
	"sync"
	"time"
)

type SpaceTimeValidator struct {
	maxSkew       time.Duration
	knownLocation map[string]string
	mu            sync.RWMutex
}

func NewSpaceTimeValidator() *SpaceTimeValidator {
	return &SpaceTimeValidator{
		maxSkew:       10 * time.Second,
		knownLocation: make(map[string]string),
	}
}

func (v *SpaceTimeValidator) senderKey(env STCEnvelope) string {
	return fmt.Sprintf("%d-%d", env.ShardID, env.NodeID)
}

func (v *SpaceTimeValidator) Validate(env STCEnvelope) error {
	now := time.Now().UnixNano()
	delta := now - env.Timestamp
	if delta < 0 {
		delta = -delta
	}
	if time.Duration(delta) > v.maxSkew {
		return fmt.Errorf("abnormal timestamp skew detected: %v", time.Duration(delta))
	}

	key := v.senderKey(env)
	v.mu.Lock()
	defer v.mu.Unlock()
	if old, ok := v.knownLocation[key]; ok {
		if old != env.Location {
			return fmt.Errorf("abnormal location mismatch detected: old=%s new=%s", old, env.Location)
		}
	} else {
		v.knownLocation[key] = env.Location
	}
	return nil
}
