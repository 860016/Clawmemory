package services

// Helper functions used across services

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
