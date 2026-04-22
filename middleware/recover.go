package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecoverSif prevents panics from crashing requests and returns a stable error payload.
func RecoverSif(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			if ctx.Writer != nil && ctx.Writer.Written() {
				ctx.Abort()
				return
			}
			ctx.Header("Content-Type", "application/json; charset=utf-8")
			ctx.String(http.StatusOK, ErrorMsg)
			ctx.Abort()
		}
	}()
	ctx.Next()
}
