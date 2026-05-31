package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"clawmemory/internal/services"
)

func main() {
	fmt.Println("=== Memory Scanner Real Test ===")
	fmt.Printf("Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	scanner := services.NewMemoryScanner()

	fmt.Println("Scanning running IDE processes...")
	items := scanner.ExtractConversations()

	fmt.Printf("\nTotal items found: %d\n\n", len(items))

	if len(items) == 0 {
		fmt.Println("No conversation items found in memory.")
		os.Exit(0)
	}

	typeCount := make(map[string]int)
	platformCount := make(map[string]int)
	for _, item := range items {
		typeCount[item.Type]++
		if item.Platform != "" {
			platformCount[item.Platform]++
		} else {
			platformCount["unknown"]++
		}
	}

	fmt.Println("=== By Type ===")
	for t, c := range typeCount {
		fmt.Printf("  %s: %d\n", t, c)
	}

	fmt.Println("\n=== By Platform ===")
	for p, c := range platformCount {
		fmt.Printf("  %s: %d\n", p, c)
	}

	fmt.Println("\n=== AI Response Items ===")
	aiCount := 0
	for _, item := range items {
		if item.Type != "ai_response" {
			continue
		}
		aiCount++
		fmt.Printf("\n--- AI Response #%d ---\n", aiCount)
		fmt.Printf("  Session:  %s\n", item.SessionID)
		fmt.Printf("  Agent:    %s\n", item.AgentID)
		if item.Thought != "" {
			preview := item.Thought
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			fmt.Printf("  Thought(%d): %s\n", len(item.Thought), preview)
		}
		if item.Content != "" {
			preview := item.Content
			if len(preview) > 500 {
				preview = preview[:500] + "..."
			}
			fmt.Printf("  Content(%d): %s\n", len(item.Content), preview)
		}
	}
	if aiCount == 0 {
		fmt.Println("  (none)")
	}

	fmt.Println("\n=== Session Items ===")
	for _, item := range items {
		if item.Type != "session" {
			continue
		}
		fmt.Printf("  %s | %s | %s\n", item.SessionID, item.Name, item.Status)
	}

	fmt.Println("\n=== User Input Items ===")
	uiCount := 0
	for _, item := range items {
		if item.Type != "user_input" {
			continue
		}
		uiCount++
		if uiCount <= 5 {
			preview := item.Content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("  %d: %s\n", uiCount, preview)
		}
	}
	fmt.Printf("  Total user inputs: %d\n", uiCount)

	data, _ := json.MarshalIndent(items, "", "  ")
	os.WriteFile("scan_results.json", data, 0644)
	fmt.Printf("\nFull results (%d items) saved to scan_results.json\n", len(items))
}
