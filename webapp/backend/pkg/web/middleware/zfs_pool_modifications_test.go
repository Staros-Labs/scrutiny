package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/analogj/scrutiny/webapp/backend/pkg/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestZFSPoolModificationGuard(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "archive", method: http.MethodPost, path: "/api/zfs/pool/1/archive"},
		{name: "unarchive", method: http.MethodPost, path: "/api/zfs/pool/1/unarchive"},
		{name: "mute", method: http.MethodPost, path: "/api/zfs/pool/1/mute"},
		{name: "unmute", method: http.MethodPost, path: "/api/zfs/pool/1/unmute"},
		{name: "label", method: http.MethodPost, path: "/api/zfs/pool/1/label"},
		{name: "delete", method: http.MethodDelete, path: "/api/zfs/pool/1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Handle(test.method, test.path, middleware.ZFSPoolModificationGuard(false), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			response := httptest.NewRecorder()
			request, err := http.NewRequest(test.method, test.path, nil)
			require.NoError(t, err)
			router.ServeHTTP(response, request)

			require.Equal(t, http.StatusForbidden, response.Code)
			require.JSONEq(t, `{"success":false,"error":"ZFS pool modifications are disabled by server configuration."}`, response.Body.String())
		})
	}
}

func TestZFSPoolModificationGuardAllowsMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/zfs/pool/1/archive", middleware.ZFSPoolModificationGuard(true), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "/api/zfs/pool/1/archive", nil)
	require.NoError(t, err)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
}
