package main

import (
	"fmt"
	"sync"
	"time"
)

type SpaceTimeValidator struct {
	mu          sync.Mutex
	knownPos    map[string]string
	allowedSkew time.Duration
	logFn       func(string, ...interface{})
}

func NewSpaceTimeValidator(logFn func(string, ...interface{})) *SpaceTimeValidator {
	return &SpaceTimeValidator{
		knownPos:    make(map[string]string),
		allowedSkew: 10 * time.Second,
		logFn:       logFn,
	}
}

func (v *SpaceTimeValidator) Validate(env SpaceTimeEnvelope) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	key := fmt.Sprintf("%d-%d", env.ShardID, env.NodeID)
	now := time.Now().Unix()
	if delta := now - env.Timestamp; delta > int64(v.allowedSkew.Seconds()) || delta < -int64(v.allowedSkew.Seconds()) {
		v.logFn("STC_ANOMALY type=time shard=%d node=%d msgTs=%d localTs=%d", env.ShardID, env.NodeID, env.Timestamp, now)
		return false
	}

	if old, ok := v.knownPos[key]; ok {
		if old != env.Location {
			v.logFn("STC_ANOMALY type=location shard=%d node=%d old=%s new=%s", env.ShardID, env.NodeID, old, env.Location)
			return false
		}
	} else {
		v.knownPos[key] = env.Location
	}

	return true
}
