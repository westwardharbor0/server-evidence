package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks if bearer token is provided when auth is on.
func AuthMiddleware(ctx *gin.Context) {
	if !apiConfig.Config.Api.Auth {
		return
	}

	auth := ctx.Request.Header.Get("Authorization")
	if auth == "" {
		ctx.String(http.StatusForbidden, "No Authorization provided")
		ctx.Abort()
		return
	}

	if strings.TrimPrefix(auth, "Bearer ") != apiConfig.Config.Api.BearerToken {
		ctx.String(http.StatusForbidden, "Authorization failed")
		ctx.Abort()
		return
	}
}

// EditableMiddleware checks if edit is enabled on PUT and DELETE endpoints.
func EditableMiddleware(ctx *gin.Context) {
	if apiConfig.Config.Machines.Readonly {
		reqMethod := ctx.Request.Method
		if reqMethod == "PUT" || reqMethod == "DELETE" {
			ctx.String(http.StatusMethodNotAllowed, "Readonly mode")
			ctx.Abort()
			return
		}
	}
}
