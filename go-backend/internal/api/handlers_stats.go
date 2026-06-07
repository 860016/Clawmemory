package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleListReports(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewDailyReportService(db)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		reports, total, err := svc.List(userID, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": reports, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleCreateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewDailyReportService(db)
		report, err := svc.Create(userID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, report)
	}
}

func handleGetReportByDate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		date := c.Param("date")
		svc := services.NewDailyReportService(db)
		report, err := svc.GetByDate(userID, date)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

func handleGenerateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			req = map[string]interface{}{}
		}
		date, _ := req["date"].(string)
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		svc := services.NewDailyReportService(db)
		report, err := svc.Generate(userID, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

func handleGetStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var memoryCount, entityCount, relationCount, projectCount int64
		logDBErr("count memories for stats", db.Model(&struct{ ID uint }{}).Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Count(&memoryCount).Error)
		logDBErr("count entities for stats", db.Model(&struct{ ID uint }{}).Table("entities").Where("user_id = ?", userID).Count(&entityCount).Error)
		logDBErr("count relations for stats", db.Model(&struct{ ID uint }{}).Table("relations").Where("user_id = ?", userID).Count(&relationCount).Error)
		logDBErr("count projects for stats", db.Model(&struct{ ID uint }{}).Table("projects").Where("user_id = ?", userID).Count(&projectCount).Error)

		layerStats := make(map[string]int64)
		rows, err := db.Raw("SELECT COALESCE(layer, 'knowledge') as layer, COUNT(*) as cnt FROM memories WHERE user_id = ? AND status != 'trashed' GROUP BY layer", userID).Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var layer string
				var cnt int64
				rows.Scan(&layer, &cnt)
				layerStats[layer] = cnt
			}
		}

		if len(layerStats) == 0 {
			layerStats["knowledge"] = 0
		}

		type RecentMemory struct {
			ID        uint      `json:"id"`
			Key       string    `json:"key"`
			Layer     string    `json:"layer"`
			CreatedAt time.Time `json:"created_at"`
		}
		var recentMemories []RecentMemory
		logDBErr("load recent memories for dashboard", db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Order("created_at desc").Limit(10).Find(&recentMemories).Error)

		recentMemoriesJson := make([]map[string]interface{}, 0)
		for _, m := range recentMemories {
			recentMemoriesJson = append(recentMemoriesJson, map[string]interface{}{
				"id":         m.ID,
				"key":        m.Key,
				"layer":      m.Layer,
				"created_at": m.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		var userCount int64
		logDBErr("count users for dashboard", db.Table("users").Count(&userCount).Error)
		passwordSet := userCount > 0

		maxMemories := int64(50000)
		settingsSvc := services.NewSettingsService(db)
		if v, err := settingsSvc.GetByKey(userID, "max_memories"); err == nil {
			if n, ok := v.(float64); ok && n > 0 {
				maxMemories = int64(n)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"memoryCount":    memoryCount,
			"entityCount":    entityCount,
			"relationCount":  relationCount,
			"projectCount":   projectCount,
			"layerStats":     layerStats,
			"recentMemories": recentMemoriesJson,
			"passwordSet":    passwordSet,
			"maxMemories":    maxMemories,
		})
	}
}

func handleGetUsageStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
		if days < 1 {
			days = 30
		}
		if days > 365 {
			days = 365
		}

		var memories []struct {
			ID         uint      `json:"id"`
			Key        string    `json:"key"`
			Layer      string    `json:"layer"`
			Source     string    `json:"source"`
			Importance float64   `json:"importance"`
			CreatedAt  time.Time `json:"created_at"`
		}
		logDBErr("load memories for export", db.Table("memories").Where("user_id = ? AND status != ?", userID, "trashed").Order("created_at desc").Find(&memories).Error)

		now := time.Now()

		dailyTrend := make([]map[string]interface{}, 0)
		for i := days - 1; i >= 0; i-- {
			date := now.AddDate(0, 0, -i).Format("2006-01-02")
			dayStart, _ := time.Parse("2006-01-02", date)
			dayEnd := dayStart.AddDate(0, 0, 1)
			count := 0
			for _, m := range memories {
				if m.CreatedAt.After(dayStart) && m.CreatedAt.Before(dayEnd) {
					count++
				}
			}
			dailyTrend = append(dailyTrend, map[string]interface{}{
				"date":  date,
				"count": count,
			})
		}

		sourceDist := make(map[string]int)
		importanceDist := make(map[string]int)
		layerDist := make(map[string]int)
		entityTypeDist := make(map[string]int)

		for _, m := range memories {
			layer := m.Layer
			if layer == "" {
				layer = "knowledge"
			}
			source := m.Source
			if source == "" {
				source = "manual"
			}
			sourceDist[source]++

			if m.Importance >= 0.7 {
				importanceDist["high"]++
			} else if m.Importance >= 0.3 {
				importanceDist["medium"]++
			} else {
				importanceDist["low"]++
			}
			layerDist[layer]++
		}

		var entityCount int64
		logDBErr("count entities for import", db.Table("entities").Where("user_id = ?", userID).Count(&entityCount).Error)

		rows, err := db.Raw("SELECT entity_type, COUNT(*) as cnt FROM entities WHERE user_id = ? GROUP BY entity_type", userID).Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var etype string
				var cnt int64
				rows.Scan(&etype, &cnt)
				entityTypeDist[etype] = int(cnt)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"dailyTrend":             dailyTrend,
			"dailyTokenTrend":        []map[string]interface{}{},
			"sourceDistribution":     sourceDist,
			"importanceDistribution": importanceDist,
			"tokenByLayer":           layerDist,
			"totalEstimatedTokens":   len(memories) * 100,
			"topAccessed":            []map[string]interface{}{},
			"operationCounts":        map[string]int{},
			"entityTypeDistribution": entityTypeDist,
			"totalMemories":          len(memories),
			"days":                   days,
		})
	}
}
