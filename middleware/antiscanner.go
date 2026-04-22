package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AntiScanner blocks obvious internet scanner traffic while keeping normal SIF traffic intact.
func AntiScanner() gin.HandlerFunc {
	denyMethods := map[string]struct{}{
		"OPTIONS":  {},
		"PROPFIND": {},
		"PRI":      {},
		"TRACE":    {},
		"CONNECT":  {},
	}

	denyPrefixes := []string{
		"/.git",
		"/phpmyadmin",
		"/phpMyAdmin",
		"/nacos",
		"/hudson",
		"/reportserver",
		"/v1/models",
		"/v2/keys",
		"/runningpods",
		"/cwbase",
		"/mcp",
		"/sse",
		"/+CSCOE+",
		"/",
	}

	denyExact := map[string]struct{}{
		"/robots.txt":                  {},
		"/favicon.ico":                 {},
		"/nice ports,/trinity.txt.bak": {},
	}

	return func(ctx *gin.Context) {
		method := strings.ToUpper(ctx.Request.Method)
		if _, blocked := denyMethods[method]; blocked {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		path := strings.ToLower(ctx.Request.URL.Path)
		if _, blocked := denyExact[path]; blocked {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		for _, p := range denyPrefixes {
			if strings.HasPrefix(path, p) {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
		}

		ctx.Next()
	}
}
