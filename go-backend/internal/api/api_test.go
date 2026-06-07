package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"clawmemory/internal/database"
	"clawmemory/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := database.Init(t.TempDir() + "\\test.db")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func closeTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	RegisterRoutes(r, db)
	return r
}

// getAuthToken creates a user via DB and logs in to get a JWT token.
func getAuthToken(t *testing.T, router *gin.Engine, db *gorm.DB) string {
	t.Helper()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	user := models.User{
		Username:  "testuser",
		Password:  string(hashedPassword),
		Role:      "admin",
		IsFounder: true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Login
	loginBody, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "testpass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d, body: %s", w.Code, w.Body.String())
	}

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token, ok := loginResp["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("login response missing access_token: %v", loginResp)
	}
	return token
}

// --- Tests ---

func TestUnauthAccess(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/memories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauth access: expected 401, got %d", w.Code)
	}
}

func TestAuthLoginAndAccess(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)
	token := getAuthToken(t, router, db)

	// Access protected route with token
	req := httptest.NewRequest(http.MethodGet, "/api/v1/memories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("auth access: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestMemoryCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)
	token := getAuthToken(t, router, db)

	// Create memory
	createBody, _ := json.Marshal(map[string]interface{}{
		"key":          "test-key",
		"value":        "test value content",
		"layer":        "core",
		"importance":   0.9,
		"memory_type":  "knowledge",
		"visibility":   "private",
		"source_agent": "test",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memories", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("create memory: expected 200/201, got %d, body: %s", w.Code, w.Body.String())
	}

	// List memories
	req = httptest.NewRequest(http.MethodGet, "/api/v1/memories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list memories: expected 200, got %d", w.Code)
	}

	var listResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)

	items, ok := listResp["items"].([]interface{})
	if !ok {
		t.Fatalf("list response missing items: %v", listResp)
	}
	if len(items) == 0 {
		t.Error("list memories: expected at least 1 item")
	}
}

func TestMemoryCreateValidation(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)
	token := getAuthToken(t, router, db)

	// Missing required "key" field
	createBody, _ := json.Marshal(map[string]interface{}{
		"value": "some value",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memories", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusCreated {
		t.Error("create memory without key should fail validation")
	}
}

func TestMemorySearch(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)
	token := getAuthToken(t, router, db)

	// Create a memory to search for
	createBody, _ := json.Marshal(map[string]interface{}{
		"key":   "search-target",
		"value": "Unique searchable content about golang programming",
		"layer": "context",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memories", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("create for search: got %d", w.Code)
	}

	// Search
	req = httptest.NewRequest(http.MethodGet, "/api/v1/memories/search?q=golang&mode=keyword", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("search: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestHealthCheck(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health check: expected 200, got %d", w.Code)
	}
}

func TestInstallStatus(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/install-status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("install status: expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["version"] == nil {
		t.Error("install status: missing version field")
	}
}

func TestEntityCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer closeTestDB(t, db)
	router := setupTestRouter(db)
	token := getAuthToken(t, router, db)

	// Create entity
	createBody, _ := json.Marshal(map[string]interface{}{
		"name":        "Go Language",
		"entity_type": "technology",
		"properties":  map[string]string{"paradigm": "concurrent"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/entities", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("create entity: expected 200/201, got %d, body: %s", w.Code, w.Body.String())
	}

	// List entities
	req = httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/entities", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("list entities: expected 200, got %d", w.Code)
	}
}
