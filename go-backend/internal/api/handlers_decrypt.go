package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleDecryptMemory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewMemoryService(db)
		memory, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if !memory.IsEncrypted {
			c.JSON(http.StatusOK, gin.H{"value": memory.Value, "encrypted": false})
			return
		}

		secretKey := services.GetEncryptionKey()
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "SECRET_KEY not configured, cannot decrypt"})
			return
		}

		encryptor, err := services.NewEncryptor(secretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to init decryptor"})
			return
		}

		decrypted, err := services.DecryptValue(encryptor, memory.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decryption failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"value": decrypted, "encrypted": true})
	}
}
