//go:build !windows && !linux && !darwin

package services

import "log"

type MemoryScanner struct{}

func NewMemoryScanner() *MemoryScanner {
	return &MemoryScanner{}
}

func (ms *MemoryScanner) ExtractConversations() []AIConversationItem {
	log.Printf("[memscan] Memory scanning is not supported on this platform")
	return nil
}
