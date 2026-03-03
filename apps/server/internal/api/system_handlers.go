package api

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleGetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"version": "0.1.0",
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}))
}

func (s *Server) handleGetSystemStats(c *gin.Context) {
	runningTunnels := s.xrayManager.GetRunningProcesses()

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"cpu":              0,
		"memory":           0,
		"disk":             0,
		"runningTunnels":   len(runningTunnels),
		"runningTunnelIds": runningTunnels,
	}))
}
