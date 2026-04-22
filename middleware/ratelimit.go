package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitRPM limits requests per minute per IP. If rpm <= 0, it is disabled.
func RateLimitRPM(rpm int) gin.HandlerFunc {
	if rpm <= 0 {
		return func(ctx *gin.Context) { ctx.Next() }
	}

	type bucket struct {
		count int
		reset time.Time
	}
	var (
		mu    sync.Mutex
		store = map[string]*bucket{}
	)

	return func(ctx *gin.Context) {
		host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
		if err != nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		now := time.Now()
		mu.Lock()
		b, ok := store[host]
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(time.Minute)}
			store[host] = b
		}
		b.count++
		allowed := b.count <= rpm
		mu.Unlock()

		if !allowed {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
