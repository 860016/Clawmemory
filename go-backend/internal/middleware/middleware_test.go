package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS_AllowLocalhost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("CORS localhost: expected origin echoed, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_BlockExternalInRelease(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("GIN_MODE", "release")
	defer os.Unsetenv("GIN_MODE")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://192.168.1.100:8080")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "http://192.168.1.100:8080" {
		t.Error("CORS: 192.168.x origin should be blocked in release mode")
	}
}

func TestCORS_AllowPrivateInDev(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Unsetenv("GIN_MODE")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://192.168.1.100:8080")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://192.168.1.100:8080" {
		t.Errorf("CORS: 192.168.x origin should be allowed in dev mode, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_CustomWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://myapp.com,https://admin.myapp.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Allowed
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://myapp.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "https://myapp.com" {
		t.Error("CORS: custom whitelist should allow myapp.com")
	}

	// Blocked
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") == "https://evil.com" {
		t.Error("CORS: custom whitelist should block evil.com")
	}
}

func TestCORS_OptionsPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("CORS preflight: expected 204, got %d", w.Code)
	}
}

func TestIsOriginAllowed_EmptyOrigin(t *testing.T) {
	if !isOriginAllowed("") {
		t.Error("empty origin should be allowed")
	}
}

func TestIsOriginAllowed_UnknownOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("GIN_MODE")

	if isOriginAllowed("https://evil.attacker.com") {
		t.Error("unknown external origin should be blocked")
	}
}
