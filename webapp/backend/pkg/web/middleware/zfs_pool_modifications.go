package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ZFSPoolModificationGuard rejects user-facing ZFS pool mutations when the server is read-only for pools.
func ZFSPoolModificationGuard(allowed bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "ZFS pool modifications are disabled by server configuration.",
			})
			return
		}

		c.Next()
	}
}
