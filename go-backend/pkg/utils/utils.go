package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
)

// GenerateID 生成唯一 ID
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetDeviceFingerprint 获取设备指纹
func GetDeviceFingerprint() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%s-%d", runtime.GOOS, runtime.GOARCH, hashString(hostname))
}

func hashString(s string) uint32 {
	var h uint32 = 5381
	for i := 0; i < len(s); i++ {
		h = ((h << 5) + h) + uint32(s[i])
	}
	return h
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
