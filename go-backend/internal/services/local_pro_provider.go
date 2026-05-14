package services

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/config"
	"clawmemory/internal/models"

	"gorm.io/gorm"
)

const (
	gracePeriodDays   = 7
	heartbeatInterval = 6 * time.Hour
	heartbeatTimeout  = 10 * time.Second
	activateTimeout   = 15 * time.Second
	deactivateTimeout = 10 * time.Second
	verifyTimeout     = 10 * time.Second
	publicKeyTimeout  = 10 * time.Second
)

type DecayEvaluator interface {
	DecayEvaluate(ctx context.Context, userID uint, isPro bool) (map[string]interface{}, error)
}

type LocalProProvider struct {
	db            *gorm.DB
	cfg           *config.Config
	aiSvc         DecayEvaluator
	mu            sync.RWMutex
	licenseKey    string
	licenseActive bool
	tier          string
	features      []string
	expiresAt     *time.Time
	fingerprint   string
	deviceName    string
	backupsDir    string
	databasePath  string
	compressCfg   map[string]interface{}
	backupCfg     struct {
		Enabled       bool
		IntervalHours int
	}
	lastHeartbeatSuccess time.Time
	heartbeatStopCh      chan struct{}
	heartbeatRunning     bool
	publicKey            *rsa.PublicKey
}

func NewLocalProProvider(db *gorm.DB, cfg *config.Config) *LocalProProvider {
	p := &LocalProProvider{
		db:  db,
		cfg: cfg,
		compressCfg: map[string]interface{}{
			"auto_compress":      false,
			"threshold":          5000,
			"level":              "light",
			"preserve_important": true,
		},
		backupCfg: struct {
			Enabled       bool
			IntervalHours int
		}{
			Enabled:       false,
			IntervalHours: 24,
		},
		tier: "oss",
	}

	p.fingerprint = generateFingerprint()
	p.deviceName = generateDeviceName()
	p.restoreFromDB()
	p.fetchPublicKey()

	if p.licenseActive {
		go p.asyncVerify()
		go p.startHeartbeat()
	}

	return p
}

func (p *LocalProProvider) SetBackupPaths(backupsDir, databasePath string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.backupsDir = backupsDir
	p.databasePath = databasePath
	os.MkdirAll(backupsDir, 0755)
}

func (p *LocalProProvider) SetAIService(svc DecayEvaluator) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.aiSvc = svc
}

func (p *LocalProProvider) IsPro() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.licenseActive && p.checkGracePeriod()
}

func (p *LocalProProvider) GetLicenseInfo() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.licenseActive {
		info := map[string]interface{}{
			"active":      true,
			"tier":        p.tier,
			"is_valid":    p.checkGracePeriod(),
			"features":    p.features,
			"key_hash":    hashKey(p.licenseKey),
			"expires_at":  nil,
			"device_slot": "",
			"fingerprint": p.fingerprint[:8],
		}
		if p.expiresAt != nil {
			info["expires_at"] = p.expiresAt.Format(time.RFC3339)
		}
		return info
	}
	return map[string]interface{}{
		"active":   false,
		"tier":     "oss",
		"is_valid": false,
		"features": []string{},
	}
}

func (p *LocalProProvider) InvalidateCache() {}

func (p *LocalProProvider) IsFeatureEnabled(feature string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.licenseActive || !p.checkGracePeriod() {
		return false
	}
	for _, f := range p.features {
		if f == feature {
			return true
		}
	}
	return false
}

func (p *LocalProProvider) SelfCheck() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.licenseActive && p.checkGracePeriod()
}

func (p *LocalProProvider) GetTier() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.licenseActive && p.checkGracePeriod() {
		return p.tier
	}
	return "oss"
}

func (p *LocalProProvider) ActivateLicense(licenseKey string) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if licenseKey == "" {
		return nil, fmt.Errorf("license key cannot be empty")
	}

	reqBody := map[string]interface{}{
		"license_key": licenseKey,
		"fingerprint": p.fingerprint,
		"version":     config.AppVersion,
		"device_name": p.deviceName,
		"os":          runtime.GOOS,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: activateTimeout}
	resp, err := client.Post(p.cfg.LicenseServerURL+"/api/v1/activate", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to connect license server: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response from license server")
	}

	if resp.StatusCode != 200 {
		msg, _ := result["message"].(string)
		if msg == "" {
			msg = "activation failed"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	valid, _ := result["valid"].(bool)
	if !valid {
		return nil, fmt.Errorf("license validation failed")
	}

	if sigRaw, ok := result["signature"].(string); ok && p.publicKey != nil {
		if err := p.verifySignature(sigRaw); err != nil {
			log.Printf("[Pro] WARNING: signature verification failed: %v", err)
		}
	}

	tier, _ := result["tier"].(string)
	if tier == "" {
		tier = "pro"
	}

	var features []string
	if fa, ok := result["features"].([]interface{}); ok {
		for _, f := range fa {
			if fs, ok := f.(string); ok {
				features = append(features, fs)
			}
		}
	}
	if len(features) == 0 {
		features = defaultProFeatures()
	}

	var expiresAt *time.Time
	if expStr, ok := result["expires_at"].(string); ok && expStr != "" {
		if t, err := time.Parse(time.RFC3339, expStr); err == nil {
			expiresAt = &t
		}
	}

	deviceSlot, _ := result["device_slot"].(string)

	p.licenseKey = licenseKey
	p.licenseActive = true
	p.tier = tier
	p.features = features
	p.expiresAt = expiresAt
	p.lastHeartbeatSuccess = time.Now()

	p.persistToDB(deviceSlot)

	go p.startHeartbeat()

	log.Printf("[Pro] License activated: tier=%s key=%s", tier, hashKey(licenseKey))

	return map[string]interface{}{
		"active":      true,
		"tier":        tier,
		"valid":       true,
		"features":    features,
		"expires_at":  result["expires_at"],
		"device_slot": deviceSlot,
		"message":     "License activated successfully",
		"key_hash":    hashKey(licenseKey),
	}, nil
}

func (p *LocalProProvider) DeactivateLicense() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.licenseKey == "" {
		return nil
	}

	reqBody := map[string]interface{}{
		"license_key": p.licenseKey,
		"fingerprint": p.fingerprint,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: deactivateTimeout}
	resp, err := client.Post(p.cfg.LicenseServerURL+"/api/v1/deactivate", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[Pro] WARNING: failed to contact server for deactivation: %v", err)
	} else {
		resp.Body.Close()
	}

	p.stopHeartbeat()

	p.licenseKey = ""
	p.licenseActive = false
	p.tier = "oss"
	p.features = nil
	p.expiresAt = nil

	p.db.Where("1 = 1").Delete(&models.License{})

	log.Printf("[Pro] License deactivated")
	return nil
}

func (p *LocalProProvider) Heartbeat() error {
	p.mu.RLock()
	key := p.licenseKey
	active := p.licenseActive
	p.mu.RUnlock()

	if !active || key == "" {
		return nil
	}

	var memoryCount int64
	p.db.Model(&models.Memory{}).Where("status != ?", "trashed").Count(&memoryCount)

	reqBody := map[string]interface{}{
		"license_key":  key,
		"memory_count": memoryCount,
		"active_users": 1,
		"version":      config.AppVersion,
		"os":           runtime.GOOS,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: heartbeatTimeout}
	resp, err := client.Post(p.cfg.LicenseServerURL+"/api/v1/heartbeat", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[Pro] heartbeat failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		p.mu.Lock()
		p.lastHeartbeatSuccess = time.Now()
		p.mu.Unlock()
	} else {
		log.Printf("[Pro] heartbeat returned status %d", resp.StatusCode)
	}

	return nil
}

func (p *LocalProProvider) DecayStats(userID uint) (map[string]interface{}, error) {
	var stats struct {
		Total    int64
		Active   int64
		Archived int64
		Trashed  int64
	}
	_ = p.db.Model(&models.Memory{}).Where("user_id = ?", userID).Count(&stats.Total).Error
	_ = p.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "active").Count(&stats.Active).Error
	_ = p.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "archived").Count(&stats.Archived).Error
	_ = p.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "trashed").Count(&stats.Trashed).Error

	var avgImportance float64
	p.db.Model(&models.Memory{}).Where("user_id = ? AND status = ?", userID, "active").
		Select("COALESCE(AVG(importance), 0)").Row().Scan(&avgImportance)

	return map[string]interface{}{
		"total":          stats.Total,
		"active":         stats.Active,
		"archived":       stats.Archived,
		"trashed":        stats.Trashed,
		"avg_importance": math.Round(avgImportance*1000) / 1000,
	}, nil
}

func (p *LocalProProvider) DecayApply(userID uint) (map[string]interface{}, error) {
	p.mu.RLock()
	svc := p.aiSvc
	p.mu.RUnlock()

	if svc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		result, err := svc.DecayEvaluate(ctx, userID, true)
		if err == nil {
			evaluations, _ := result["evaluations"].([]map[string]interface{})
			archived, trashed, kept := 0, 0, 0
			for _, ev := range evaluations {
				action, _ := ev["action"].(string)
				switch action {
				case "archive":
					archived++
				case "delete":
					trashed++
				default:
					kept++
				}
			}
			result["processed"] = len(evaluations)
			result["archived"] = archived
			result["trashed"] = trashed
			result["adjusted"] = kept
			result["algorithm"] = "ai_decay_v1"
			result["mode"] = "ai"
			return result, nil
		}
		log.Printf("[Pro] AI decay evaluate failed, falling back to DecayService: %v", err)
	}

	decaySvc := NewDecayService(p.db)
	result, err := decaySvc.ApplyDecay(userID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = map[string]interface{}{}
	}
	result["mode"] = "local_fallback"
	return result, nil
}

func (p *LocalProProvider) PruneSuggest(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	suggestions := []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < 0.2 {
			suggestions = append(suggestions, map[string]interface{}{
				"id":                 m.ID,
				"key":                m.Key,
				"layer":              m.Layer,
				"importance":         m.Importance,
				"decayed_importance": m.Importance * 0.7,
				"reason":             "low_importance",
			})
		}
	}

	return map[string]interface{}{
		"mode":        "local",
		"suggestions": suggestions,
		"total":       len(suggestions),
	}, nil
}

func (p *LocalProProvider) ConflictScan(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	keyMap := make(map[string][]models.Memory)
	for _, m := range memories {
		keyMap[m.Key] = append(keyMap[m.Key], m)
	}

	conflicts := []map[string]interface{}{}

	for key, mems := range keyMap {
		if len(mems) > 1 {
			items := make([]map[string]interface{}, 0, len(mems))
			for _, m := range mems {
				items = append(items, map[string]interface{}{
					"id":         m.ID,
					"value":      truncateStr(m.Value, 100),
					"importance": m.Importance,
					"updated_at": m.UpdatedAt,
				})
			}
			conflicts = append(conflicts, map[string]interface{}{
				"key":      key,
				"count":    len(mems),
				"severity": "exact_duplicate",
				"memories": items,
			})
		}
	}

	tagMemories := make(map[string][]models.Memory)
	for _, m := range memories {
		if m.Tags == "" {
			continue
		}
		for _, tag := range strings.Split(m.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagMemories[tag] = append(tagMemories[tag], m)
			}
		}
	}

	for tag, mems := range tagMemories {
		if len(mems) < 2 || len(mems) > 50 {
			continue
		}
		for i := 0; i < len(mems); i++ {
			for j := i + 1; j < len(mems); j++ {
				if mems[i].Key == mems[j].Key {
					continue
				}
				sim := jaccardSimilarity(mems[i].Value, mems[j].Value)
				if sim >= 0.5 {
					conflicts = append(conflicts, map[string]interface{}{
						"key":        fmt.Sprintf("%s vs %s", mems[i].Key, mems[j].Key),
						"count":      2,
						"severity":   "similar_content",
						"similarity": sim,
						"tag":        tag,
						"memories": []map[string]interface{}{
							{"id": mems[i].ID, "value": truncateStr(mems[i].Value, 100), "importance": mems[i].Importance},
							{"id": mems[j].ID, "value": truncateStr(mems[j].Value, 100), "importance": mems[j].Importance},
						},
					})
				}
			}
		}
	}

	sort.Slice(conflicts, func(i, j int) bool {
		si, _ := conflicts[i]["severity"].(string)
		sj, _ := conflicts[j]["severity"].(string)
		prio := map[string]int{"exact_duplicate": 3, "similar_content": 2, "potential_conflict": 1}
		return prio[si] > prio[sj]
	})

	return map[string]interface{}{
		"mode":      "local",
		"conflicts": conflicts,
		"total":     len(conflicts),
	}, nil
}

func (p *LocalProProvider) ConflictResolve(userID uint, conflictIndex int, strategy string) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	keyMap := make(map[string][]models.Memory)
	for _, m := range memories {
		keyMap[m.Key] = append(keyMap[m.Key], m)
	}

	keys := make([]string, 0, len(keyMap))
	for k := range keyMap {
		if len(keyMap[k]) > 1 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	if conflictIndex < 0 || conflictIndex >= len(keys) {
		return nil, fmt.Errorf("conflict index %d out of range (0-%d)", conflictIndex, len(keys)-1)
	}

	targetKey := keys[conflictIndex]
	duplicates := keyMap[targetKey]

	resolved := 0
	switch strategy {
	case "keep_newest":
		sort.Slice(duplicates, func(i, j int) bool {
			return duplicates[i].UpdatedAt.After(duplicates[j].UpdatedAt)
		})
		for _, d := range duplicates[1:] {
			if err := p.db.Model(&models.Memory{}).Where("id = ?", d.ID).
				Update("status", "archived").Error; err == nil {
				resolved++
			}
		}
	case "keep_important":
		sort.Slice(duplicates, func(i, j int) bool {
			return duplicates[i].Importance > duplicates[j].Importance
		})
		for _, d := range duplicates[1:] {
			if err := p.db.Model(&models.Memory{}).Where("id = ?", d.ID).
				Update("status", "archived").Error; err == nil {
				resolved++
			}
		}
	case "merge":
		merged := duplicates[0]
		for _, d := range duplicates[1:] {
			merged.Value = merged.Value + "\n" + d.Value
			if err := p.db.Model(&models.Memory{}).Where("id = ?", d.ID).
				Update("status", "archived").Error; err == nil {
				resolved++
			}
		}
		p.db.Model(&models.Memory{}).Where("id = ?", merged.ID).
			Updates(map[string]interface{}{"value": merged.Value, "updated_at": time.Now()})
	default:
		return nil, fmt.Errorf("unknown strategy: %s (use: keep_newest, keep_important, merge)", strategy)
	}

	return map[string]interface{}{
		"mode":     "local",
		"key":      targetKey,
		"strategy": strategy,
		"resolved": resolved,
	}, nil
}

func (p *LocalProProvider) CompressPreview(userID uint, level string) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status != ?", userID, "trashed").Limit(5000).Find(&memories).Error

	threshold := 0.3
	switch level {
	case "light":
		threshold = 0.2
	case "medium":
		threshold = 0.35
	case "heavy":
		threshold = 0.5
	default:
		threshold = 0.2
	}

	preview := []map[string]interface{}{}
	for _, m := range memories {
		if m.Importance < threshold {
			preview = append(preview, map[string]interface{}{
				"memory_id":  m.ID,
				"key":        m.Key,
				"value_len":  len(m.Value),
				"importance": m.Importance,
				"action":     "archive",
			})
		}
	}

	return map[string]interface{}{
		"mode":      "local",
		"level":     level,
		"threshold": threshold,
		"preview":   preview,
		"total":     len(preview),
	}, nil
}

func (p *LocalProProvider) CompressApply(userID uint, level string) (map[string]interface{}, error) {
	preview, err := p.CompressPreview(userID, level)
	if err != nil {
		return nil, err
	}

	previewItems, _ := preview["preview"].([]map[string]interface{})
	archived := 0
	for _, item := range previewItems {
		if id, ok := item["memory_id"]; ok {
			var memoryID uint
			switch v := id.(type) {
			case uint:
				memoryID = v
			case float64:
				memoryID = uint(v)
			case int:
				memoryID = uint(v)
			}
			if err := p.db.Model(&models.Memory{}).Where("id = ? AND user_id = ?", memoryID, userID).
				Update("status", "archived").Error; err == nil {
				archived++
			}
		}
	}

	return map[string]interface{}{
		"mode":     "local",
		"level":    level,
		"archived": archived,
		"total":    len(previewItems),
	}, nil
}

func (p *LocalProProvider) CompressConfig(userID uint) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cfg := make(map[string]interface{})
	for k, v := range p.compressCfg {
		cfg[k] = v
	}
	return cfg, nil
}

func (p *LocalProProvider) SetCompressConfig(userID uint, cfg map[string]interface{}) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range cfg {
		p.compressCfg[k] = v
	}
	return map[string]interface{}{
		"updated": true,
		"config":  p.compressCfg,
	}, nil
}

func (p *LocalProProvider) TokenRoute(message string, contextLength int) (map[string]interface{}, error) {
	length := len([]rune(message))
	sentenceCount := 0
	for _, ch := range message {
		if ch == '。' || ch == '.' || ch == '？' || ch == '?' || ch == '！' || ch == '!' {
			sentenceCount++
		}
	}
	if sentenceCount == 0 {
		sentenceCount = 1
	}
	avgSentenceLen := float64(length) / float64(sentenceCount)

	technicalTerms := []string{"算法", "架构", "API", "数据库", "函数", "模型", "优化", "部署", "系统", "框架", "协议", "接口"}
	techCount := 0
	for _, term := range technicalTerms {
		if strings.Contains(message, term) {
			techCount++
		}
	}

	score := 1
	if length > 100 {
		score++
	}
	if avgSentenceLen > 30 {
		score++
	}
	if techCount >= 2 {
		score++
	}

	complexityLabel := "simple"
	layer := "core"
	strategy := "direct"
	switch score {
	case 1:
		complexityLabel = "simple"
		layer = "core"
		strategy = "direct"
	case 2:
		complexityLabel = "medium"
		layer = "core"
		strategy = "keyword_priority"
	case 3:
		complexityLabel = "complex"
		layer = "semantic"
		strategy = "semantic_priority"
	case 4:
		complexityLabel = "extreme"
		layer = "summary"
		strategy = "full_context"
	}

	tokenEstimate := length / 4
	if tokenEstimate == 0 {
		tokenEstimate = 1
	}

	return map[string]interface{}{
		"mode":              "local",
		"token_estimate":    tokenEstimate,
		"recommended_layer": layer,
		"strategy":          strategy,
		"complexity":        complexityLabel,
		"complexity_score":  score,
		"technical_terms":   techCount,
		"avg_sentence_len":  math.Round(avgSentenceLen*10) / 10,
	}, nil
}

func (p *LocalProProvider) TokenStats(userID uint) (map[string]interface{}, error) {
	var totalMemories int64
	_ = p.db.Model(&models.Memory{}).Where("user_id = ? AND status != ?", userID, "trashed").Count(&totalMemories).Error

	var totalEntities int64
	_ = p.db.Model(&models.Entity{}).Where("user_id = ?", userID).Count(&totalEntities).Error

	var totalRelations int64
	_ = p.db.Model(&models.Relation{}).Where("user_id = ?", userID).Count(&totalRelations).Error

	estimatedTokens := totalMemories*200 + totalEntities*50 + totalRelations*30

	return map[string]interface{}{
		"mode":            "local",
		"memory_tokens":   totalMemories * 200,
		"entity_tokens":   totalEntities * 50,
		"relation_tokens": totalRelations * 30,
		"total_tokens":    estimatedTokens,
		"memory_count":    totalMemories,
		"entity_count":    totalEntities,
		"relation_count":  totalRelations,
	}, nil
}

func (p *LocalProProvider) EvolutionInsights(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	layerCount := make(map[string]int)
	sourceCount := make(map[string]int)
	for _, m := range memories {
		layerCount[m.Layer]++
		if m.Source != "" {
			sourceCount[m.Source]++
		}
	}

	return map[string]interface{}{
		"mode":         "local",
		"total":        len(memories),
		"layer_stats":  layerCount,
		"source_stats": sourceCount,
	}, nil
}

func (p *LocalProProvider) EvolutionDiscover(userID uint) (map[string]interface{}, error) {
	var relations []models.Relation
	_ = p.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error

	typeCount := make(map[string]int)
	for _, r := range relations {
		typeCount[r.RelationType]++
	}

	discoveries := []map[string]interface{}{}
	for rtype, count := range typeCount {
		if count >= 2 {
			discoveries = append(discoveries, map[string]interface{}{
				"relation_type": rtype,
				"count":         count,
				"confidence":    math.Min(float64(count)/10.0, 1.0),
			})
		}
	}

	return map[string]interface{}{
		"mode":        "local",
		"discoveries": discoveries,
		"total":       len(discoveries),
	}, nil
}

func (p *LocalProProvider) EvolutionInfer(userID uint) (map[string]interface{}, error) {
	var entities []models.Entity
	_ = p.db.Where("user_id = ?", userID).Limit(5000).Find(&entities).Error

	inferences := []map[string]interface{}{}
	for _, e := range entities {
		if e.Confidence >= 0.7 {
			inferences = append(inferences, map[string]interface{}{
				"entity_id":   e.ID,
				"entity_name": e.Name,
				"confidence":  e.Confidence,
				"reason":      "high_importance_entity",
			})
		}
	}

	return map[string]interface{}{
		"mode":       "local",
		"inferences": inferences,
		"total":      len(inferences),
	}, nil
}

func (p *LocalProProvider) EvolutionImportance(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	importanceBuckets := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}
	for _, m := range memories {
		switch {
		case m.Importance >= 0.8:
			importanceBuckets["critical"]++
		case m.Importance >= 0.5:
			importanceBuckets["high"]++
		case m.Importance >= 0.3:
			importanceBuckets["medium"]++
		default:
			importanceBuckets["low"]++
		}
	}

	return map[string]interface{}{
		"mode":    "local",
		"buckets": importanceBuckets,
		"total":   len(memories),
	}, nil
}

func (p *LocalProProvider) EvolutionPrefetch(userID uint, context string) (map[string]interface{}, error) {
	var memories []models.Memory
	escaped := escapeLike(context)
	_ = p.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(200).Find(&memories).Error

	results := make([]map[string]interface{}, 0, len(memories))
	for _, m := range memories {
		results = append(results, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"importance": m.Importance,
			"layer":      m.Layer,
		})
	}

	return map[string]interface{}{
		"mode":    "local",
		"context": context,
		"results": results,
		"total":   len(results),
	}, nil
}

func (p *LocalProProvider) ReinforceMemory(userID uint, memoryID uint) (map[string]interface{}, error) {
	var memory models.Memory
	if err := p.db.Where("id = ? AND user_id = ?", memoryID, userID).First(&memory).Error; err != nil {
		return nil, fmt.Errorf("memory not found: %w", err)
	}

	newImportance := math.Min(memory.Importance+0.1, 1.0)
	if err := p.db.Model(&memory).Updates(map[string]interface{}{
		"importance":   newImportance,
		"access_count": gorm.Expr("access_count + 1"),
		"updated_at":   time.Now(),
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to reinforce memory: %w", err)
	}

	return map[string]interface{}{
		"mode":           "local",
		"memory_id":      memoryID,
		"old_importance": memory.Importance,
		"new_importance": newImportance,
		"reinforced":     true,
	}, nil
}

func (p *LocalProProvider) AutoGraph(userID uint, overwrite bool) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ?", userID, "active").Limit(5000).Find(&memories).Error

	var relations []models.Relation
	_ = p.db.Where("user_id = ?", userID).Limit(5000).Find(&relations).Error

	existingPairs := make(map[string]bool)
	for _, r := range relations {
		key := fmt.Sprintf("%d-%s-%d", r.SourceID, r.RelationType, r.TargetID)
		existingPairs[key] = true
	}

	created := 0

	keyMemories := make(map[string][]models.Memory)
	for _, m := range memories {
		parts := strings.SplitN(m.Key, ":", 2)
		if len(parts) == 2 {
			keyMemories[parts[0]] = append(keyMemories[parts[0]], m)
		}
	}

	for _, mems := range keyMemories {
		if len(mems) < 2 {
			continue
		}
		for i := 0; i < len(mems); i++ {
			for j := i + 1; j < len(mems); j++ {
				pairKey := fmt.Sprintf("%d-same_topic-%d", mems[i].ID, mems[j].ID)
				if existingPairs[pairKey] && !overwrite {
					continue
				}
				rel := models.Relation{
					SourceID:     mems[i].ID,
					TargetID:     mems[j].ID,
					RelationType: "same_topic",
					UserID:       userID,
					Weight:       0.6,
				}
				if err := p.db.Create(&rel).Error; err == nil {
					created++
				}
			}
		}
	}

	tagMemories := make(map[string][]models.Memory)
	for _, m := range memories {
		if m.Tags == "" {
			continue
		}
		for _, tag := range strings.Split(m.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagMemories[tag] = append(tagMemories[tag], m)
			}
		}
	}

	for _, mems := range tagMemories {
		if len(mems) < 2 || len(mems) > 100 {
			continue
		}
		for i := 0; i < len(mems); i++ {
			for j := i + 1; j < len(mems); j++ {
				if mems[i].Key == mems[j].Key {
					continue
				}
				pairKey := fmt.Sprintf("%d-shared_tags-%d", mems[i].ID, mems[j].ID)
				if existingPairs[pairKey] && !overwrite {
					continue
				}
				rel := models.Relation{
					SourceID:     mems[i].ID,
					TargetID:     mems[j].ID,
					RelationType: "shared_tags",
					UserID:       userID,
					Weight:       0.4,
				}
				if err := p.db.Create(&rel).Error; err == nil {
					created++
				}
			}
		}
	}

	var entities []models.Entity
	_ = p.db.Where("user_id = ?", userID).Limit(1000).Find(&entities).Error

	entityMemories := make(map[uint][]uint)
	for _, e := range entities {
		if e.SourceMemoryID != nil {
			entityMemories[e.ID] = append(entityMemories[e.ID], *e.SourceMemoryID)
		}
	}

	for _, memIDs := range entityMemories {
		if len(memIDs) < 2 {
			continue
		}
		for i := 0; i < len(memIDs); i++ {
			for j := i + 1; j < len(memIDs); j++ {
				pairKey := fmt.Sprintf("%d-shared_entity-%d", memIDs[i], memIDs[j])
				if existingPairs[pairKey] && !overwrite {
					continue
				}
				rel := models.Relation{
					SourceID:     memIDs[i],
					TargetID:     memIDs[j],
					RelationType: "shared_entity",
					UserID:       userID,
					Weight:       0.5,
				}
				if err := p.db.Create(&rel).Error; err == nil {
					created++
				}
			}
		}
	}

	return map[string]interface{}{
		"mode":        "local",
		"created":     created,
		"total_pairs": len(existingPairs),
	}, nil
}

func (p *LocalProProvider) AIExtract(userID uint) (map[string]interface{}, error) {
	var memories []models.Memory
	_ = p.db.Where("user_id = ? AND status != ?", userID, "trashed").Select("value").Limit(5000).Find(&memories).Error

	extracted := 0
	for _, m := range memories {
		if m.Value == "" {
			continue
		}
		var existingCount int64
		p.db.Model(&models.Entity{}).Where("user_id = ? AND name = ?", userID, truncateStr(m.Key, 200)).Count(&existingCount)
		if existingCount > 0 {
			continue
		}
		entity := models.Entity{
			Name:       truncateStr(m.Key, 200),
			EntityType: "concept",
			Confidence: m.Importance,
			UserID:     userID,
		}
		if err := p.db.Create(&entity).Error; err == nil {
			extracted++
		}
	}

	return map[string]interface{}{
		"mode":      "local",
		"extracted": extracted,
		"scanned":   len(memories),
	}, nil
}

func (p *LocalProProvider) BackupSchedule(userID uint) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"enabled":        p.backupCfg.Enabled,
		"interval_hours": p.backupCfg.IntervalHours,
	}, nil
}

func (p *LocalProProvider) SetBackupSchedule(userID uint, enabled bool, intervalHours int) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if intervalHours < 1 {
		intervalHours = 24
	}
	if intervalHours > 720 {
		intervalHours = 720
	}
	p.backupCfg.Enabled = enabled
	p.backupCfg.IntervalHours = intervalHours
	return map[string]interface{}{
		"enabled":        enabled,
		"interval_hours": intervalHours,
		"updated":        true,
	}, nil
}

func (p *LocalProProvider) SmartLoad(userID uint, context string) (map[string]interface{}, error) {
	var memories []models.Memory
	escaped := escapeLike(context)
	_ = p.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(200).Find(&memories).Error

	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Importance > memories[j].Importance
	})

	if len(memories) > 50 {
		memories = memories[:50]
	}

	results := make([]map[string]interface{}, 0, len(memories))
	for _, m := range memories {
		results = append(results, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"value":      truncateStr(m.Value, 200),
			"importance": m.Importance,
			"layer":      m.Layer,
		})
	}

	return map[string]interface{}{
		"mode":    "local",
		"context": context,
		"results": results,
		"total":   len(results),
	}, nil
}

func generateFingerprint() string {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	data := fmt.Sprintf("%s|%s|%s|%s", hostname, runtime.GOOS, username, "clawmemory")
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func generateDeviceName() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Unknown"
	}
	return fmt.Sprintf("%s/%s", runtime.GOOS, hostname)
}

func (p *LocalProProvider) checkGracePeriod() bool {
	if p.lastHeartbeatSuccess.IsZero() {
		return true
	}
	sinceLast := time.Since(p.lastHeartbeatSuccess)
	if sinceLast > time.Duration(gracePeriodDays)*24*time.Hour {
		log.Printf("[Pro] grace period expired (%.0f days since last verification)", sinceLast.Hours()/24)
		return false
	}
	return true
}

func (p *LocalProProvider) restoreFromDB() {
	var license models.License
	if err := p.db.Where("status = ?", "active").First(&license).Error; err != nil {
		return
	}

	if license.ExpiresAt != nil && license.ExpiresAt.Before(time.Now()) {
		p.db.Model(&license).Update("status", "expired")
		log.Printf("[Pro] stored license expired, deactivating")
		return
	}

	p.licenseKey = license.LicenseKey
	p.licenseActive = true
	p.tier = license.Tier
	p.deviceName = license.DeviceName
	p.expiresAt = license.ExpiresAt

	if license.Features != "" {
		var feats []string
		if err := json.Unmarshal([]byte(license.Features), &feats); err == nil {
			p.features = feats
		}
	}
	if len(p.features) == 0 {
		p.features = defaultProFeatures()
	}

	p.lastHeartbeatSuccess = time.Now()

	log.Printf("[Pro] restored license from DB: tier=%s key=%s", p.tier, hashKey(p.licenseKey))
}

func (p *LocalProProvider) persistToDB(deviceSlot string) {
	featuresJSON, _ := json.Marshal(p.features)

	var existing models.License
	err := p.db.Where("1 = 1").First(&existing).Error

	now := time.Now()
	record := models.License{
		LicenseKey:        p.licenseKey,
		Tier:              p.tier,
		Status:            "active",
		DeviceFingerprint: p.fingerprint,
		DeviceName:        p.deviceName,
		ExpiresAt:         p.expiresAt,
		DeviceSlot:        deviceSlot,
		Features:          string(featuresJSON),
		ActivatedAt:       &now,
	}

	if err == nil {
		p.db.Model(&existing).Updates(map[string]interface{}{
			"license_key":        record.LicenseKey,
			"tier":               record.Tier,
			"status":             record.Status,
			"device_fingerprint": record.DeviceFingerprint,
			"device_name":        record.DeviceName,
			"expires_at":         record.ExpiresAt,
			"device_slot":        record.DeviceSlot,
			"features":           record.Features,
			"activated_at":       record.ActivatedAt,
		})
	} else {
		p.db.Create(&record)
	}
}

func (p *LocalProProvider) fetchPublicKey() {
	client := &http.Client{Timeout: publicKeyTimeout}
	resp, err := client.Get(p.cfg.LicenseServerURL + "/api/v1/public-key")
	if err != nil {
		log.Printf("[Pro] WARNING: failed to fetch public key: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[Pro] WARNING: public-key endpoint returned %d", resp.StatusCode)
		return
	}

	keyBytes, _ := io.ReadAll(resp.Body)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		log.Printf("[Pro] WARNING: failed to decode PEM public key")
		return
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("[Pro] WARNING: failed to parse public key: %v", err)
		return
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		log.Printf("[Pro] WARNING: public key is not RSA")
		return
	}

	p.publicKey = rsaPub

	keyDir := filepath.Dir(p.cfg.RSAPublicKeyPath)
	os.MkdirAll(keyDir, 0755)
	os.WriteFile(p.cfg.RSAPublicKeyPath, keyBytes, 0644)

	log.Printf("[Pro] public key fetched and cached")
}

func (p *LocalProProvider) verifySignature(sigRaw string) error {
	sigBytes, err := base64.StdEncoding.DecodeString(sigRaw)
	if err != nil {
		return fmt.Errorf("base64 decode failed: %w", err)
	}

	var sigData struct {
		Data      string `json:"data"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(sigBytes, &sigData); err != nil {
		return fmt.Errorf("json unmarshal failed: %w", err)
	}

	sig, err := base64.StdEncoding.DecodeString(sigData.Signature)
	if err != nil {
		return fmt.Errorf("signature base64 decode failed: %w", err)
	}

	hashed := sha256.Sum256([]byte(sigData.Data))
	if err := rsa.VerifyPKCS1v15(p.publicKey, crypto.SHA256, hashed[:], sig); err != nil {
		return fmt.Errorf("RSA verification failed: %w", err)
	}

	return nil
}

func (p *LocalProProvider) asyncVerify() {
	time.Sleep(5 * time.Second)

	p.mu.RLock()
	key := p.licenseKey
	p.mu.RUnlock()

	if key == "" {
		return
	}

	reqBody := map[string]interface{}{
		"license_key": key,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: verifyTimeout}
	resp, err := client.Post(p.cfg.LicenseServerURL+"/api/v1/verify", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[Pro] async verify: server unreachable, entering grace period")
		p.mu.Lock()
		p.lastHeartbeatSuccess = time.Now()
		p.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[Pro] async verify: license invalid on server, deactivating")
		p.mu.Lock()
		p.licenseActive = false
		p.tier = "oss"
		p.features = nil
		p.mu.Unlock()
		p.db.Where("1 = 1").Delete(&models.License{})
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if valid, _ := result["valid"].(bool); valid {
		p.mu.Lock()
		p.lastHeartbeatSuccess = time.Now()
		if tier, ok := result["tier"].(string); ok && tier != "" {
			p.tier = tier
		}
		p.mu.Unlock()
		log.Printf("[Pro] async verify: license confirmed valid")
	} else {
		log.Printf("[Pro] async verify: license not valid, deactivating")
		p.mu.Lock()
		p.licenseActive = false
		p.tier = "oss"
		p.features = nil
		p.mu.Unlock()
		p.db.Where("1 = 1").Delete(&models.License{})
	}
}

func (p *LocalProProvider) startHeartbeat() {
	p.mu.Lock()
	if p.heartbeatRunning {
		p.mu.Unlock()
		return
	}
	p.heartbeatRunning = true
	p.heartbeatStopCh = make(chan struct{})
	p.mu.Unlock()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	_ = p.Heartbeat()

	for {
		select {
		case <-ticker.C:
			_ = p.Heartbeat()
		case <-p.heartbeatStopCh:
			return
		}
	}
}

func (p *LocalProProvider) stopHeartbeat() {
	if p.heartbeatStopCh != nil {
		close(p.heartbeatStopCh)
	}
	p.heartbeatRunning = false
	p.heartbeatStopCh = nil
}

func defaultProFeatures() []string {
	return []string{
		"ai_extract", "auto_graph", "unlimited_graph", "auto_decay",
		"decay_report", "prune_suggest", "reinforce", "conflict_scan",
		"conflict_merge", "smart_router", "token_stats", "wiki",
		"auto_backup", "compress", "evolution",
	}
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:8])
}
