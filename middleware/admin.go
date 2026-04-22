package middleware

import (
	"net"
	"net/http"	
	"strings"

	"github.com/gin-gonic/gin"	
)

// AdminGate blocks admin routes when disabled or when IP is not allowlisted.
func AdminGate(enabled bool, allowIPs []string) gin.HandlerFunc {
	allow := map[string]struct{}{}
	for _, ip := range allowIPs {
		allow[strings.TrimSpace(ip)] = struct{}{}
	}
	return func(ctx *gin.Context) {
		if !enabled {
			ctx.Status(http.StatusNotFound)
			ctx.Abort()
			return
		}
		if len(allow) > 0 {
			host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
			if err != nil {
				ctx.Status(http.StatusForbidden)
				ctx.Abort()
				return
			}
			if _, ok := allow[host]; !ok {
				ctx.Status(http.StatusForbidden)
				ctx.Abort()
				return
			}
		}
		ctx.Next()
	}
}
