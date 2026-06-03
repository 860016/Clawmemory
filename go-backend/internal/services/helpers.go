package services

import (
	"encoding/json"
	"log"
	"strings"
)

func getString(data map[string]interface{}, key, defaultValue string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return defaultValue
}

func getFloat(data map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return defaultValue
}

func getBool(data map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := data[key].(bool); ok {
		return v
	}
	return defaultValue
}

func getInt(data map[string]interface{}, key string, defaultValue int) int {
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return defaultValue
}

func logDBErr(context string, err error) {
	if err != nil {
		log.Printf("[DB] %s: %v", context, err)
	}
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func parseMemoryTags(tags string) []string {
	if tags == "" || tags == "[]" {
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(tags), &result); err == nil {
		return result
	}
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func tokenSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setB := make(map[string]struct{}, len(b))
	for _, t := range b {
		setB[strings.ToLower(t)] = struct{}{}
	}
	matches := 0
	for _, t := range a {
		if _, ok := setB[strings.ToLower(t)]; ok {
			matches++
		}
	}
	return float64(matches) / float64(len(a))
}
