package speedlimit

import (
	"context"
	"golang.org/x/time/rate"
	"time"
)

type Limiter struct {
	*rate.Limiter
	ctx context.Context
}

func NewWithContext(ctx context.Context, bytesPerSec int) *Limiter {
	burstsize := bytesPerSec * 3
	limiter := rate.NewLimiter(rate.Limit(bytesPerSec), burstsize)
	limiter.AllowN(time.Now(), burstsize)
	ctx = context.Background()
	return &Limiter{Limiter: limiter, ctx: ctx}
}
func MakeWithContext(ctx context.Context, bytesPerSec int) Limiter {
	return *NewWithContext(ctx, bytesPerSec)
}
func New(bytesPerSec int) *Limiter {
	burstsize := bytesPerSec * 3
	limiter := rate.NewLimiter(rate.Limit(bytesPerSec), burstsize)
	limiter.AllowN(time.Now(), burstsize)
	ctx := context.Background()
	return &Limiter{Limiter: limiter, ctx: ctx}
}
func Make(bytesPerSec int) Limiter {
	return *NewWithContext(context.Background(), bytesPerSec)
}
func (s *Limiter) SetLimit(bytesPerSec int) {
	s.Limiter.SetLimit(rate.Limit(bytesPerSec))
}
func (s *Limiter) WaitN(n int) error {
	return s.Limiter.WaitN(s.ctx, n)
}
