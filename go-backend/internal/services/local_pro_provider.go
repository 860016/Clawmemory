package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type LocalProProvider struct {
	db            *gorm.DB
	mu            sync.RWMutex
	licenseKey    string
	licenseActive bool
	backupsDir    string
	databasePath  string
	compressCfg   map[string]interface{}
	backupCfg     struct {
		Enabled       bool
		IntervalHours int
	}
}

func NewLocalProProvider(db *gorm.DB) *LocalProProvider {
	return &LocalProProvider{
		db: db,
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
	}
}

func (p *LocalProProvider) SetBackupPaths(backupsDir, databasePath string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.backupsDir = backupsDir
	p.databasePath = databasePath
	os.MkdirAll(backupsDir, 0755)
}

func (p *LocalProProvider) IsPro() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.licenseActive
}

func (p *LocalProProvider) GetLicenseInfo() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.licenseActive {
		return map[string]interface{}{
			"active":   true,
			"tier":     "pro",
			"is_valid": true,
			"features": []string{"decay", "conflict", "compress", "token_route", "evolution", "reinforce", "auto_graph", "ai_extract", "backup_schedule"},
			"key_hash": hashKey(p.licenseKey),
		}
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
	return p.licenseActive
}

func (p *LocalProProvider) SelfCheck() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.licenseActive
}

func (p *LocalProProvider) GetTier() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.licenseActive {
		return "pro"
	}
	return "oss"
}

func (p *LocalProProvider) ActivateLicense(licenseKey string) (map[string]interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if licenseKey == "" {
		return nil, fmt.Errorf("license key cannot be empty")
	}

	if len(licenseKey) < 16 {
		return nil, fmt.Errorf("invalid license key format")
	}

	p.licenseKey = licenseKey
	p.licenseActive = true

	log.Printf("[Pro] License activated: %s", hashKey(licenseKey))

	return map[string]interface{}{
		"active":   true,
		"tier":     "pro",
		"message":  "License activated successfully",
		"key_hash": hashKey(licenseKey),
	}, nil
}

func (p *LocalProProvider) DeactivateLicense() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.licenseKey = ""
	p.licenseActive = false

	log.Printf("[Pro] License deactivated")
	return nil
}

func (p *LocalProProvider) Heartbeat() error {
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
	var lowMemories []models.Memory
	_ = p.db.Where("user_id = ? AND status = ? AND importance < ?", userID, "active", 0.2).
		Limit(50).Find(&lowMemories).Error

	archived := 0
	for _, m := range lowMemories {
		if err := p.db.Model(&models.Memory{}).Where("id = ?", m.ID).
			Updates(map[string]interface{}{"status": "archived", "updated_at": time.Now()}).Error; err == nil {
			archived++
		}
	}

	return map[string]interface{}{
		"mode":     "local",
		"archived": archived,
		"checked":  len(lowMemories),
	}, nil
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
				"memories": items,
			})
		}
	}

	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i]["count"].(int) > conflicts[j]["count"].(int)
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
	tokenEstimate := len(message) / 4
	if tokenEstimate == 0 {
		tokenEstimate = 1
	}

	layer := "core"
	if tokenEstimate > 8000 {
		layer = "summary"
	} else if tokenEstimate > 2000 {
		layer = "semantic"
	}

	return map[string]interface{}{
		"mode":              "local",
		"token_estimate":    tokenEstimate,
		"recommended_layer": layer,
		"strategy":          "keyword_priority",
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
				pairKey := fmt.Sprintf("%d-related-%d", mems[i].ID, mems[j].ID)
				if existingPairs[pairKey] && !overwrite {
					continue
				}
				rel := models.Relation{
					SourceID:     mems[i].ID,
					TargetID:     mems[j].ID,
					RelationType: "related",
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

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:8])
}
