package relay

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	*rate.Limiter
	ctx context.Context
}

func NewWithContext(ctx context.Context, bytesPerSec int) *Limiter {
	burstsize := bytesPerSec * 1
	limiter := rate.NewLimiter(rate.Limit(bytesPerSec), burstsize)
	limiter.AllowN(time.Now(), burstsize)
	ctx = context.Background()
	return &Limiter{Limiter: limiter, ctx: ctx}
}
func MakeWithContext(ctx context.Context, bytesPerSec int) Limiter {
	return *NewWithContext(ctx, bytesPerSec)
}
func NewSpeedLimiter(bytesPerSec int) *Limiter {
	return NewWithContext(context.Background(), bytesPerSec)
}
func MakeSpeedLimiter(bytesPerSec int) Limiter {
	return *NewWithContext(context.Background(), bytesPerSec)
}
func (s *Limiter) SetLimit(bytesPerSec int) {
	burstsize := bytesPerSec * 1
	s.Limiter = rate.NewLimiter(rate.Limit(bytesPerSec), burstsize)
}
func (s *Limiter) WaitN(n int) (err error) {
	if err = s.Limiter.WaitN(s.ctx, n); err == nil {
		return
	} else {
		sleepDuration := n * 1000 / s.Limiter.Burst()
		time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
		return nil
	}
	return
}
