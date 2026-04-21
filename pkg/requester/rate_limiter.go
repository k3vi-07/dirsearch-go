package requester

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(maxRate int) *RateLimiter {
	if maxRate <= 0 {
		return &RateLimiter{
			limiter: rate.NewLimiter(rate.Inf, 0),
		}
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(maxRate), maxRate),
	}
}

// Wait 等待令牌
func (r *RateLimiter) Wait(ctx context.Context) error {
	return r.limiter.Wait(ctx)
}

// Allow 检查是否允许
func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}

// SetRate 设置速率
func (r *RateLimiter) SetRate(newRate int) {
	if newRate <= 0 {
		r.limiter.SetLimit(rate.Inf)
		r.limiter.SetBurst(0)
	} else {
		r.limiter.SetLimit(rate.Limit(newRate))
		r.limiter.SetBurst(newRate)
	}
}

// GetRate 获取当前速率
func (r *RateLimiter) GetRate() rate.Limit {
	return r.limiter.Limit()
}
