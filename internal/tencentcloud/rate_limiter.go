package tencentcloud

import (
	"sync"
	"time"
)

// RateLimiter is a simple token-bucket style limiter.
type RateLimiter struct {
	maxCalls      int
	periodSeconds float64
	mu            sync.Mutex
	timestamps    []time.Time
}

func NewRateLimiter(maxCalls int, periodSeconds float64) *RateLimiter {
	return &RateLimiter{
		maxCalls:      maxCalls,
		periodSeconds: periodSeconds,
	}
}

func (r *RateLimiter) WaitIfNeeded() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Duration(r.periodSeconds * float64(time.Second)))
	filtered := make([]time.Time, 0, len(r.timestamps))
	for _, t := range r.timestamps {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	r.timestamps = filtered

	if len(r.timestamps) >= r.maxCalls {
		sleep := time.Duration(r.periodSeconds*float64(time.Second)) - now.Sub(r.timestamps[0])
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
	r.timestamps = append(r.timestamps, time.Now())
}
