package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleMCPConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		apiKeySvc := services.NewAPIKeyService(db)
		keys, err := apiKeySvc.List(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var rawKey string
		if len(keys) == 0 {
			_, rk, err := apiKeySvc.Create(userID, "mcp-server")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key: " + err.Error()})
				return
			}
			rawKey = rk
		}

		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, host)

		if rawKey == "" {
			for _, k := range keys {
				if k.Name == "mcp-server" {
					rawKey = k.KeyPrefix + "••••••••"
					break
				}
			}
			if rawKey == "" && len(keys) > 0 {
				rawKey = keys[0].KeyPrefix + "••••••••"
			}
		}

		mcpServerConfig := map[string]interface{}{
			"command": "npx",
			"args":    []string{"-y", "clawmemory-mcp"},
			"env": map[string]string{
				"CLAWMEMORY_BASE_URL": baseURL,
				"CLAWMEMORY_API_KEY":  rawKey,
			},
		}

		cursorConfig := map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"clawmemory": mcpServerConfig,
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"base_url":    baseURL,
			"api_key":     rawKey,
			"has_new_key": len(keys) == 0,
			"configs": map[string]interface{}{
				"cursor": map[string]interface{}{
					"label":      "Cursor",
					"configPath": "~/.cursor/mcp.json",
					"config":     cursorConfig,
				},
				"claude_desktop": map[string]interface{}{
					"label":      "Claude Desktop",
					"configPath": "~/AppData/Roaming/Claude/claude_desktop_config.json",
					"config":     cursorConfig,
				},
				"windsurf": map[string]interface{}{
					"label":      "Windsurf",
					"configPath": "~/.windsurf/mcp.json",
					"config":     cursorConfig,
				},
				"trae": map[string]interface{}{
					"label":      "Trae",
					"configPath": "~/.trae/mcp.json",
					"config":     cursorConfig,
				},
			},
		})
	}
}
