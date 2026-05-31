package main

import (
	"fmt"
	"time"

	"clawmemory/internal/services"
)

func main() {
	fmt.Println("=== Memory Scanner Diagnostics ===")
	fmt.Printf("Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	scanner := services.NewMemoryScanner()

	pids := scanner.FindTraeRendererPIDs()
	fmt.Printf("Found %d renderer PIDs\n\n", len(pids))

	for _, rp := range pids {
		fmt.Printf("--- PID %d (%s, %d MB) ---\n", rp.PID, rp.ToolName, rp.MemMB)

		thoughtResults := scanner.ScanForPatternRaw(rp.PID, `"thought"`, 50, 500000, 100)
		fmt.Printf("  'thought' raw results: %d\n", len(thoughtResults))

		planResults := scanner.ScanForPatternRaw(rp.PID, "plan_item", 100, 500000, 100)
		fmt.Printf("  'plan_item' raw results: %d\n", len(planResults))

		sessionResults := scanner.ScanForPatternRaw(rp.PID, "session_status", 500, 500000, 100)
		fmt.Printf("  'session_status' raw results: %d\n", len(sessionResults))

		inputResults := scanner.ScanForPattern(rp.PID, "inputText", 100, 200000, 100)
		fmt.Printf("  'inputText' results: %d\n", len(inputResults))

		if len(thoughtResults) > 0 {
			fmt.Printf("\n  First thought result length: %d\n", len(thoughtResults[0]))
			preview := thoughtResults[0]
			if len(preview) > 200 {
				preview = preview[:200]
			}
			fmt.Printf("  Preview: %s\n", preview)
		}

		fmt.Println()
	}
}
