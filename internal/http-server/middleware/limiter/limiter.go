package limiter

import (
	"net"
	"net/http"
	"github.com/git-217/golang-api/internal/lib/limiter"
)

func RateLimiterMiddleware(ipLimiter *limiter.IPRateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Invalid IP", http.StatusInternalServerError)
				return
			}

			if !ipLimiter.GetLimiter(ip).Allow() {
				http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
