package system_agent

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type AgentController struct{}

func (c *AgentController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/system/agent", c.DownloadAgent)
}

// DownloadAgent
// @Summary Download agent binary
// @Description Download the databasus-agent binary for the specified architecture
// @Tags system/agent
// @Produce octet-stream
// @Param arch query string true "Target architecture" Enums(amd64, arm64)
// @Success 200 {file} binary
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /system/agent [get]
func (c *AgentController) DownloadAgent(ctx *gin.Context) {
	arch := ctx.Query("arch")
	if arch != "amd64" && arch != "arm64" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "arch must be amd64 or arm64"})
		return
	}

	binaryName := "databasus-agent-linux-" + arch
	binaryPath := filepath.Join("agent-binaries", binaryName)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		ctx.JSON(
			http.StatusNotFound,
			gin.H{"error": "agent binary not found for architecture: " + arch},
		)
		return
	}

	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename=databasus-agent")
	ctx.File(binaryPath)
}
