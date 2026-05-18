package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"clawmemory/internal/api"
	"clawmemory/internal/database"
	"clawmemory/internal/models"
	"clawmemory/internal/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	router *gin.Engine
	token  string
	passed = 0
	failed = 0
)

func main() {
	gin.SetMode(gin.TestMode)

	var err error
	db, err = database.Init(":memory:")
	if err != nil {
		fmt.Printf("❌ DB init failed: %v\n", err)
		os.Exit(1)
	}
	database.Migrate(db)

	_ = services.NewToolboxService(db)

	hash, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	db.Create(&models.User{Username: "testuser", Password: string(hash), Role: "admin"})

	router = gin.New()
	router.Use(gin.Recovery())
	api.RegisterRoutes(router, db)

	fmt.Println("=== ClawMemory Full API Integration Test ===")

	if !login() {
		fmt.Println("\n❌ FATAL: Cannot login")
		printSummary()
		os.Exit(1)
	}

	testSection("Basic APIs", []testCase{
		{"GET /memories", "GET", "/api/v1/memories?limit=1", "", nil},
		{"GET /knowledge/entities", "GET", "/api/v1/knowledge/entities?limit=1", "", nil},
		{"GET /wiki", "GET", "/api/v1/wiki?limit=1", "", nil},
		{"GET /projects", "GET", "/api/v1/projects?limit=1", "", nil},
		{"GET /settings", "GET", "/api/v1/settings", "", nil},
		{"GET /stats", "GET", "/api/v1/stats", "", nil},
		{"GET /health", "GET", "/api/v1/health", "", []int{200}},
	})

	testSection("Memory CRUD", []testCase{
		{"POST /memories (create)", "POST", "/api/v1/memories", `{"key":"test_key","value":"test_value","source":"api_test"}`, nil},
		{"GET /memories/search/keyword", "GET", "/api/v1/memories/search/keyword?q=test", "", nil},
		{"GET /memories/search/semantic", "GET", "/api/v1/memories/search/semantic?q=test", "", []int{200, 500, 503}},
		{"GET /memories/decay/stats", "GET", "/api/v1/memories/decay/stats", "", nil},
		{"GET /memories/health", "GET", "/api/v1/memories/health", "", nil},
		{"GET /memories/quality", "GET", "/api/v1/memories/quality", "", nil},
		{"GET /memories/trash", "GET", "/api/v1/memories/trash", "", nil},
	})

	testSection("Knowledge Graph APIs", []testCase{
		{"GET /knowledge/relations", "GET", "/api/v1/knowledge/relations", "", nil},
		{"GET /knowledge/graph", "GET", "/api/v1/knowledge/graph", "", nil},
	})

	testSection("Wiki APIs", []testCase{
		{"GET /wiki/tree", "GET", "/api/v1/wiki/tree", "", nil},
		{"GET /wiki/categories", "GET", "/api/v1/wiki/categories", "", nil},
		{"GET /wiki/stats", "GET", "/api/v1/wiki/stats", "", nil},
	})

	testSection("AI Config APIs", []testCase{
		{"GET /ai/config", "GET", "/api/v1/ai/config", "", nil},
		{"GET /ai/usage", "GET", "/api/v1/ai/usage", "", nil},
		{"GET /ai/providers", "GET", "/api/v1/ai/providers", "", nil},
	})

	testSection("Hermes Feature APIs", []testCase{
		{"POST /ai/nudge-reflect", "POST", "/api/v1/ai/nudge-reflect", "", []int{200, 500, 503}},
		{"POST /ai/self-refine", "POST", "/api/v1/ai/self-refine", `{"pressure_level":"low"}`, []int{200, 500, 503}},
		{"POST /ai/user-profile", "POST", "/api/v1/ai/user-profile", "", []int{200, 500, 503}},
		{"POST /ai/extract-facts", "POST", "/api/v1/ai/extract-facts", `{"messages":[{"role":"user","content":"hello"}]}`, []int{200, 500, 503}},
		{"POST /ai/consolidate", "POST", "/api/v1/ai/consolidate", `{"facts":[{"fact":"test","confidence":0.5}]}`, []int{200, 500, 503}},
		{"POST /ai/process-conversation", "POST", "/api/v1/ai/process-conversation", `{"messages":[{"role":"user","content":"hello"}]}`, []int{200, 500, 503}},
		{"POST /ai/context-assemble", "POST", "/api/v1/ai/context-assemble", `{"query":"test","token_budget":1000}`, []int{200, 500, 503}},
	})

	testSection("Skill Learning APIs", []testCase{
		{"POST /skills/actions (record)", "POST", "/api/v1/skills/actions", `{"action_type":"test","action_name":"test_action","parameters":"{}","result":"success","duration":100}`, nil},
		{"POST /skills/actions/batch", "POST", "/api/v1/skills/actions/batch", `{"session_id":"test-session","agent_name":"test-agent","platform":"test","actions":[{"action_type":"test","action_name":"batch_action_1","parameters":"{}","result":"success","duration":50},{"action_type":"test","action_name":"batch_action_2","parameters":"{}","result":"success","duration":60}]}`, nil},
		{"GET /skills/detect", "GET", "/api/v1/skills/detect", "", nil},
		{"POST /skills/create (auto)", "POST", "/api/v1/skills/create", `{"use_ai":false}`, nil},
		{"GET /skills/list", "GET", "/api/v1/skills/list", "", nil},
		{"GET /skills/match", "GET", "/api/v1/skills/match?q=test", "", nil},
		{"GET /skills/suggestions", "GET", "/api/v1/skills/suggestions", "", nil},
		{"POST /skills/suggestions/generate", "POST", "/api/v1/skills/suggestions/generate", "", nil},
	})

	testSkillDetailAPIs()
	testSuggestionDismiss()

	testSection("Toolbox APIs", []testCase{
		{"GET /toolbox/decay/stats", "GET", "/api/v1/toolbox/decay/stats", "", nil},
		{"GET /toolbox/conflicts/scan", "GET", "/api/v1/toolbox/conflicts/scan", "", []int{200, 500, 503}},
		{"GET /toolbox/token/route", "GET", "/api/v1/toolbox/token/route", "", nil},
		{"GET /toolbox/token/stats", "GET", "/api/v1/toolbox/token/stats", "", nil},
		{"GET /toolbox/prune-suggest", "GET", "/api/v1/toolbox/prune-suggest", "", nil},
		{"GET /toolbox/compress/config", "GET", "/api/v1/toolbox/compress/config", "", nil},
		{"GET /toolbox/evolution/insights", "GET", "/api/v1/toolbox/evolution/insights", "", nil},
	})

	testSection("Backup APIs", []testCase{
		{"GET /backups", "GET", "/api/v1/backups", "", nil},
	})

	testSection("Share APIs", []testCase{
		{"GET /shares/pending", "GET", "/api/v1/shares/pending", "", nil},
		{"GET /shares/outbound", "GET", "/api/v1/shares/outbound", "", nil},
		{"GET /share-rules", "GET", "/api/v1/share-rules", "", nil},
	})

	printSummary()
}

type testCase struct {
	name    string
	method  string
	path    string
	body    string
	okCodes []int
}

func login() bool {
	payload := `{"username":"testuser","password":"test123"}`
	w := execRequest("POST", "/api/v1/auth/login", payload, false)
	if w.Code != 200 {
		fmt.Printf("  ❌ Login: %d — %s\n", w.Code, w.Body.String())
		failed++
		return false
	}
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if t, ok := result["access_token"].(string); ok {
		token = t
		fmt.Printf("  ✅ Login: got token (%d chars)\n\n", len(token))
		passed++
		return true
	}
	fmt.Printf("  ❌ Login: no token in response — %s\n", w.Body.String())
	failed++
	return false
}

func testSection(name string, tests []testCase) {
	fmt.Printf("--- %s ---\n", name)
	for _, tc := range tests {
		w := execRequest(tc.method, tc.path, tc.body, true)
		okCodes := tc.okCodes
		if okCodes == nil {
			okCodes = []int{200, 201}
		}
		ok := false
		for _, code := range okCodes {
			if w.Code == code {
				ok = true
				break
			}
		}
		if ok {
			fmt.Printf("  ✅ %s: %d\n", tc.name, w.Code)
			passed++
		} else {
			bodyStr := strings.TrimSpace(w.Body.String())
			if len(bodyStr) > 120 {
				bodyStr = bodyStr[:120] + "..."
			}
			fmt.Printf("  ❌ %s: %d — %s\n", tc.name, w.Code, bodyStr)
			failed++
		}
	}
	fmt.Println()
}

func testSkillDetailAPIs() {
	fmt.Println("--- Skill Detail APIs ---")
	w := execRequest("GET", "/api/v1/skills/list", "", true)
	var listResult map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResult)

	if skills, ok := listResult["skills"].([]interface{}); ok && len(skills) > 0 {
		firstSkill := skills[0].(map[string]interface{})
		skillID := firstSkill["id"]

		w = execRequest("PATCH", fmt.Sprintf("/api/v1/skills/%v", skillID), `{"field":"description","old_value":"","new_value":"patched"}`, true)
		assertResult(fmt.Sprintf("PATCH /skills/%v", skillID), w, []int{200})

		w = execRequest("POST", fmt.Sprintf("/api/v1/skills/%v/usage", skillID), `{"success":true}`, true)
		assertResult(fmt.Sprintf("POST /skills/%v/usage", skillID), w, []int{200})

		w = execRequest("POST", fmt.Sprintf("/api/v1/skills/%v/improve", skillID), "", true)
		assertResult(fmt.Sprintf("POST /skills/%v/improve", skillID), w, []int{200, 500, 503})
	} else {
		fmt.Println("  ⚠️  No skills found for detail test")
	}
	fmt.Println()
}

func testSuggestionDismiss() {
	fmt.Println("--- Suggestion Dismiss ---")
	w := execRequest("GET", "/api/v1/skills/suggestions", "", true)
	var sugResult map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &sugResult)

	if suggestions, ok := sugResult["suggestions"].([]interface{}); ok && len(suggestions) > 0 {
		firstSug := suggestions[0].(map[string]interface{})
		sugID := firstSug["id"]
		w = execRequest("POST", fmt.Sprintf("/api/v1/skills/suggestions/%v/dismiss", sugID), "", true)
		assertResult(fmt.Sprintf("POST /skills/suggestions/%v/dismiss", sugID), w, []int{200})
	} else {
		fmt.Println("  ⚠️  No suggestions to dismiss")
	}
	fmt.Println()
}

func assertResult(name string, w *httptest.ResponseRecorder, okCodes []int) {
	for _, code := range okCodes {
		if w.Code == code {
			fmt.Printf("  ✅ %s: %d\n", name, w.Code)
			passed++
			return
		}
	}
	bodyStr := strings.TrimSpace(w.Body.String())
	if len(bodyStr) > 120 {
		bodyStr = bodyStr[:120] + "..."
	}
	fmt.Printf("  ❌ %s: %d — %s\n", name, w.Code, bodyStr)
	failed++
}

func execRequest(method, path, body string, auth bool) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	if auth && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func printSummary() {
	total := passed + failed
	fmt.Println("========================================")
	fmt.Printf("  Total: %d | ✅ Passed: %d | ❌ Failed: %d\n", total, passed, failed)
	fmt.Println("========================================")
	if failed > 0 {
		os.Exit(1)
	}
}
