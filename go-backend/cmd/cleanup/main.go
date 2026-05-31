package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clawmemory/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func isReadableContent(content string) bool {
	runes := []rune(content)
	if len(runes) == 0 {
		return false
	}
	printable := 0
	cjkCount := 0
	consecutiveCJK := 0
	maxConsecutiveCJK := 0
	for _, r := range runes {
		if r >= 0x20 && r < 0x7F || r == '\n' || r == '\r' || r == '\t' {
			printable++
			consecutiveCJK = 0
		} else if r >= 0x4E00 && r <= 0x9FFF {
			printable++
			cjkCount++
			consecutiveCJK++
			if consecutiveCJK > maxConsecutiveCJK {
				maxConsecutiveCJK = consecutiveCJK
			}
		} else if r >= 0x3000 && r <= 0x303F || r >= 0xFF00 && r <= 0xFFEF {
			printable++
			consecutiveCJK = 0
		} else {
			consecutiveCJK = 0
		}
	}
	ratio := float64(printable) / float64(len(runes))
	if ratio < 0.7 {
		return false
	}
	if maxConsecutiveCJK >= 3 || cjkCount >= 5 {
		return true
	}
	v8HeapPatterns := []string{
		"MemoryScanner", "ScanForPattern", "cleanPrintable",
		"extractThought", "extractAIResponse", "regexp.MustCompile",
		"inherit", "prototype", "undefined", "constructor",
		"__proto__", "webpack", "chunkId", "moduleId",
	}
	for _, pat := range v8HeapPatterns {
		if strings.Contains(content, pat) {
			return false
		}
	}
	jsFileCount := strings.Count(content, ".js")
	if jsFileCount >= 3 {
		return false
	}
	goFileCount := strings.Count(content, ".go")
	if goFileCount >= 3 {
		return false
	}
	tsCount := strings.Count(content, ".ts")
	if tsCount >= 3 {
		return false
	}
	spaceCount := strings.Count(content, "   ")
	if spaceCount > len(runes)/20 {
		return false
	}
	blankLines := strings.Count(content, "\n\n\n")
	if blankLines >= 2 {
		return false
	}
	return true
}

func main() {
	fmt.Println("=== Cleanup Garbage Memories ===")
	fmt.Printf("Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	dbPath := filepath.Join("data", "clawmemory.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("Database not found at %s\n", dbPath)
		return
	}

	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}

	var total int64
	db.Model(&models.Memory{}).Count(&total)
	fmt.Printf("Total memories before cleanup: %d\n", total)

	batchSize := 500
	offset := 0
	deleted := 0
	kept := 0
	sampleGarbage := []string{}

	for {
		var memories []models.Memory
		result := db.Limit(batchSize).Offset(offset).Find(&memories)
		if result.Error != nil || len(memories) == 0 {
			break
		}

		for _, m := range memories {
			if !isReadableContent(m.Value) {
				if deleted < 5 {
					sampleGarbage = append(sampleGarbage, fmt.Sprintf("  ID=%d Key=%s Value=%.60s...", m.ID, m.Key, m.Value))
				}
				db.Delete(&m)
				deleted++
			} else {
				kept++
			}
		}

		offset += batchSize
		if offset%5000 == 0 {
			fmt.Printf("  Processed %d/%d (deleted %d, kept %d)\n", offset, total, deleted, kept)
		}
	}

	fmt.Printf("\nCleanup complete!\n")
	fmt.Printf("  Deleted: %d garbage memories\n", deleted)
	fmt.Printf("  Kept: %d valid memories\n", kept)

	if len(sampleGarbage) > 0 {
		fmt.Printf("\nSample garbage items deleted:\n")
		for _, s := range sampleGarbage {
			fmt.Println(s)
		}
	}

	var afterTotal int64
	db.Model(&models.Memory{}).Count(&afterTotal)
	fmt.Printf("\nTotal memories after cleanup: %d\n", afterTotal)
}
