//go:build !solution

package ratelimit

import (
	"context"
	"errors"
	"time"
)

// Limiter is precise rate limiter with context support.
type Limiter struct {
	sema      chan struct{}
	releaseTs chan time.Time
	interval  time.Duration
	stopped   bool
	abortJob  chan struct{}
}

var ErrStopped = errors.New("limiter stopped")

// NewLimiter returns limiter that throttles rate of successful Acquire() calls
// to maxSize events at any given interval.
func NewLimiter(maxCount int, interval time.Duration) *Limiter {
	limiter := &Limiter{
		sema:      make(chan struct{}, maxCount),
		releaseTs: make(chan time.Time, maxCount),
		interval:  interval,
		stopped:   false,
		abortJob:  make(chan struct{}, 1),
	}
	go limiter.releaseJob()
	return limiter
}

func (l *Limiter) Acquire(ctx context.Context) error {
	if l.stopped {
		return ErrStopped
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case l.sema <- struct{}{}:
		l.releaseTs <- time.Now().Add(l.interval)
	}
	return nil
}

func (l *Limiter) Stop() {
	l.stopped = true
	l.abortJob <- struct{}{}
	close(l.releaseTs)
}

func (l *Limiter) releaseJob() {
	for ts := range l.releaseTs {
		select {
		case <-time.After(time.Until(ts)):
			<-l.sema
		case <-l.abortJob:
			return
		}
	}
}
