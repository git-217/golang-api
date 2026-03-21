package limiter

import (
	"psql_crud/internal/config"
	"sync"
)

type IPRateLimiter struct {
	limiters   map[string]*RateLimiter
	mutex      sync.RWMutex
	maxTokens  float64
	refillRate float64
}

func NewIPRateLimiter(cfg *config.Config) *IPRateLimiter {
	return &IPRateLimiter{
		limiters:   make(map[string]*RateLimiter),
		maxTokens:  cfg.RateLimiter.MaxTokens,
		refillRate: cfg.RateLimiter.RefillRate,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *RateLimiter {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = NewRateLimiter(i.maxTokens, i.refillRate)
		i.limiters[ip] = limiter
	}

	return limiter
}
