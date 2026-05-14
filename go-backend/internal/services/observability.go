package services

import (
	"clawmemory/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

type StructuredLogger struct {
	mu      sync.Mutex
	file    *os.File
	db      *gorm.DB
	level   string
	service string
}

var (
	globalLogger *StructuredLogger
	loggerOnce   sync.Once
)

func InitLogger(logDir, level, service string) *StructuredLogger {
	loggerOnce.Do(func() {
		os.MkdirAll(logDir, 0755)

		logFile := filepath.Join(logDir, fmt.Sprintf("clawmemory-%s.log", time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Failed to open log file: %v, using stdout only", err)
		}

		globalLogger = &StructuredLogger{
			file:    f,
			level:   level,
			service: service,
		}
	})
	return globalLogger
}

func GetLogger() *StructuredLogger {
	if globalLogger == nil {
		return InitLogger("logs", "info", "clawmemory")
	}
	return globalLogger
}

func (l *StructuredLogger) SetDB(db *gorm.DB) {
	l.db = db
}

type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Service   string      `json:"service"`
	Message   string      `json:"message"`
	Caller    string      `json:"caller,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func (l *StructuredLogger) Info(msg string, data ...interface{}) {
	if l.shouldLog("info") {
		l.log("INFO", msg, data...)
	}
}

func (l *StructuredLogger) Warn(msg string, data ...interface{}) {
	if l.shouldLog("warn") {
		l.log("WARN", msg, data...)
	}
}

func (l *StructuredLogger) Error(msg string, data ...interface{}) {
	if l.shouldLog("error") {
		l.log("ERROR", msg, data...)
	}
}

func (l *StructuredLogger) Debug(msg string, data ...interface{}) {
	if l.shouldLog("debug") {
		l.log("DEBUG", msg, data...)
	}
}

func (l *StructuredLogger) shouldLog(level string) bool {
	levels := map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}
	currentLevel, ok := levels[l.level]
	if !ok {
		currentLevel = 1
	}
	msgLevel, ok := levels[level]
	if !ok {
		msgLevel = 1
	}
	return msgLevel >= currentLevel
}

func (l *StructuredLogger) log(level, msg string, data ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	caller := fmt.Sprintf("%s:%d", filepath.Base(file), line)

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Service:   l.service,
		Message:   msg,
		Caller:    caller,
	}

	if len(data) > 0 {
		if m, ok := data[0].(map[string]interface{}); ok {
			entry.Data = m
		}
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[%s] %s: marshal error: %v", level, msg, err)
		return
	}

	log.Println(string(jsonBytes))

	if l.file != nil {
		l.mu.Lock()
		l.file.WriteString(string(jsonBytes) + "\n")
		l.mu.Unlock()
	}

	if l.db != nil && (level == "ERROR" || level == "WARN") {
		l.persistLog(entry)
	}
}

func (l *StructuredLogger) persistLog(entry LogEntry) {
	if l.db == nil {
		return
	}
	dataStr := ""
	if entry.Data != nil {
		if b, err := json.Marshal(entry.Data); err == nil {
			dataStr = string(b)
		}
	}
	sysLog := &models.SystemLog{
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Service:   entry.Service,
		Message:   entry.Message,
		Caller:    entry.Caller,
		Data:      dataStr,
	}
	l.db.Create(sysLog)
}

type MetricsCollector struct {
	mu            sync.RWMutex
	requestCounts map[string]*atomic.Int64
	errorCounts   map[string]*atomic.Int64
	responseTimes map[string]*responseTimeTracker
	startTime     time.Time
	db            *gorm.DB
}

type responseTimeTracker struct {
	count int64
	total time.Duration
	max   time.Duration
	min   time.Duration
}

var (
	globalMetrics *MetricsCollector
	metricsOnce   sync.Once
)

func InitMetrics(db *gorm.DB) *MetricsCollector {
	metricsOnce.Do(func() {
		globalMetrics = &MetricsCollector{
			requestCounts: make(map[string]*atomic.Int64),
			errorCounts:   make(map[string]*atomic.Int64),
			responseTimes: make(map[string]*responseTimeTracker),
			startTime:     time.Now(),
			db:            db,
		}
	})
	return globalMetrics
}

func GetMetrics() *MetricsCollector {
	return globalMetrics
}

func (m *MetricsCollector) RecordRequest(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.requestCounts[endpoint]; !ok {
		m.requestCounts[endpoint] = &atomic.Int64{}
	}
	m.requestCounts[endpoint].Add(1)
}

func (m *MetricsCollector) RecordError(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.errorCounts[endpoint]; !ok {
		m.errorCounts[endpoint] = &atomic.Int64{}
	}
	m.errorCounts[endpoint].Add(1)
}

func (m *MetricsCollector) RecordResponseTime(endpoint string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.responseTimes[endpoint]; !ok {
		m.responseTimes[endpoint] = &responseTimeTracker{
			min: duration,
		}
	}
	tracker := m.responseTimes[endpoint]
	tracker.count++
	tracker.total += duration
	if duration > tracker.max {
		tracker.max = duration
	}
	if duration < tracker.min {
		tracker.min = duration
	}
}

func (m *MetricsCollector) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requests := make(map[string]int64)
	for k, v := range m.requestCounts {
		requests[k] = v.Load()
	}

	errors := make(map[string]int64)
	for k, v := range m.errorCounts {
		errors[k] = v.Load()
	}

	times := make(map[string]interface{})
	for k, v := range m.responseTimes {
		var avg time.Duration
		if v.count > 0 {
			avg = v.total / time.Duration(v.count)
		}
		times[k] = map[string]interface{}{
			"count": v.count,
			"avg":   avg.String(),
			"max":   v.max.String(),
			"min":   v.min.String(),
		}
	}

	return map[string]interface{}{
		"uptime_seconds": time.Since(m.startTime).Seconds(),
		"requests":       requests,
		"errors":         errors,
		"response_times": times,
	}
}

type HealthChecker struct {
	db *gorm.DB
}

func NewHealthChecker(db *gorm.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    float64                `json:"uptime_seconds"`
	Checks    map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func (h *HealthChecker) Check() HealthStatus {
	status := HealthStatus{
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Checks:    make(map[string]CheckResult),
	}

	allHealthy := true

	start := time.Now()
	sqlDB, err := h.db.DB()
	if err != nil {
		status.Checks["database"] = CheckResult{Status: "unhealthy", Detail: err.Error()}
		allHealthy = false
	} else if err := sqlDB.Ping(); err != nil {
		status.Checks["database"] = CheckResult{Status: "unhealthy", Detail: err.Error()}
		allHealthy = false
	} else {
		latency := time.Since(start)
		status.Checks["database"] = CheckResult{
			Status:  "healthy",
			Latency: latency.String(),
		}
	}

	start = time.Now()
	chromaSvc := NewChromaDBService(h.db)
	if chromaSvc.IsAvailable() {
		latency := time.Since(start)
		status.Checks["chromadb"] = CheckResult{
			Status:  "healthy",
			Latency: latency.String(),
		}
	} else {
		status.Checks["chromadb"] = CheckResult{
			Status: "degraded",
			Detail: "ChromaDB not available, using fallback search",
		}
	}

	provider := GetProProvider()
	if provider != nil && provider.IsPro() {
		status.Checks["license"] = CheckResult{
			Status: "healthy",
			Detail: fmt.Sprintf("tier: %s", provider.GetTier()),
		}
	} else {
		status.Checks["license"] = CheckResult{
			Status: "healthy",
			Detail: "OSS mode",
		}
	}

	embSvc := GetEmbeddingService()
	if embSvc != nil {
		stats := embSvc.CacheStats()
		status.Checks["embedding"] = CheckResult{
			Status: "healthy",
			Detail: fmt.Sprintf("cache_size: %v", stats["size"]),
		}
	} else {
		status.Checks["embedding"] = CheckResult{
			Status: "degraded",
			Detail: "using FNV hash fallback",
		}
	}

	if allHealthy {
		status.Status = "healthy"
	} else {
		status.Status = "unhealthy"
	}

	metrics := GetMetrics()
	if metrics != nil {
		if stats := metrics.GetStats(); stats != nil {
			if uptime, ok := stats["uptime_seconds"].(float64); ok {
				status.Uptime = uptime
			}
		}
	}
	if status.Uptime == 0 {
		metrics := GetMetrics()
		if metrics != nil {
			stats := metrics.GetStats()
			if u, ok := stats["uptime_seconds"].(float64); ok {
				status.Uptime = u
			}
		}
	}

	return status
}
