package api

import (
	"clawmemory/internal/middleware"
	"clawmemory/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func handleListProjects(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}
		status := c.Query("status")
		category := c.Query("category")

		svc := services.NewProjectService(db)
		projects, total, err := svc.List(userID, page, size, status, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": projects, "total": total, "page": page, "size": size, "pages": (total + int64(size) - 1) / int64(size)})
	}
}

func handleGetProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		project, err := svc.Get(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

func handleCreateProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		project, err := svc.Create(userID, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, project)
	}
}

func handleUpdateProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		project, err := svc.Update(userID, uint(id), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

func handleDeleteProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		if err := svc.Delete(userID, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleProjectNotes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		projectID, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		notes, err := svc.GetNotes(userID, uint(projectID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": notes})
	}
}

func handleAddProjectNote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		projectID, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		note, err := svc.AddNote(userID, uint(projectID), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, note)
	}
}

func handleUpdateProjectNote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		noteID, ok := parseIDParam(c, "noteId")
		if !ok {
			return
		}
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := services.NewProjectService(db)
		note, err := svc.UpdateNote(userID, uint(noteID), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, note)
	}
}

func handleDeleteProjectNote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		noteID, ok := parseIDParam(c, "noteId")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		if err := svc.DeleteNote(userID, uint(noteID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func handleProjectCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewProjectService(db)
		categories, err := svc.GetCategories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

func handleProjectExtractMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}

		svc := services.NewProjectService(db)
		extracted, err := svc.ExtractFromMemories(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"extracted": extracted})
	}
}

func handleProjectContext(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter required"})
			return
		}

		svc := services.NewProjectService(db)
		context, err := svc.GetContextForOpenClaw(userID, name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"context": context})
	}
}

func handleProjectSearch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		query := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		svc := services.NewProjectService(db)
		projects, err := svc.Search(userID, query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": projects})
	}
}

func handleProjectDiscoverFromMemories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		svc := services.NewProjectService(db)
		result, err := svc.DiscoverFromMemories(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"projects": result})
	}
}

func handleProjectGenerateWiki(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, ok := parseIDParam(c, "id")
		if !ok {
			return
		}
		svc := services.NewProjectService(db)
		count, err := svc.GenerateWikiFromMemories(userID, uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"wiki_entries_generated": count})
	}
}
