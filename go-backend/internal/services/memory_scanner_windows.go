//go:build windows

package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32DLL        = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess    = kernel32DLL.NewProc("OpenProcess")
	procReadProcessMem = kernel32DLL.NewProc("ReadProcessMemory")
	procVirtualQueryEx = kernel32DLL.NewProc("VirtualQueryEx")
	procCloseHandle    = kernel32DLL.NewProc("CloseHandle")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	MEM_COMMIT                = 0x1000
	PAGE_READONLY             = 0x02
	PAGE_READWRITE            = 0x04
	PAGE_WRITECOPY            = 0x08
	PAGE_EXECUTE_READ         = 0x20
	PAGE_EXECUTE_READWRITE    = 0x40
	PAGE_EXECUTE_WRITECOPY    = 0x80
	READABLE                  = PAGE_READONLY | PAGE_READWRITE | PAGE_WRITECOPY | PAGE_EXECUTE_READ | PAGE_EXECUTE_READWRITE | PAGE_EXECUTE_WRITECOPY
)

type memoryBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

type rendererProcessInfo struct {
	PID        int
	MemMB      int64
	Platform   string
	ToolName   string
	IsRenderer bool
}

type MemoryScanner struct{}

func NewMemoryScanner() *MemoryScanner {
	return &MemoryScanner{}
}

func (ms *MemoryScanner) findAllRendererPIDs() []rendererProcessInfo {
	var allRenderers []rendererProcessInfo

	for _, tool := range supportedIDETolls {
		renderers := ms.findRendererPIDsForProcess(tool.ProcessName, tool.DisplayName, tool.Platform)
		allRenderers = append(allRenderers, renderers...)
	}

	return allRenderers
}

func (ms *MemoryScanner) findRendererPIDsForProcess(processName string, displayName string, platform string) []rendererProcessInfo {
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("name='%s'", processName), "get", "processid,workingsetsize,commandline")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	var result []rendererProcessInfo
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "ProcessId") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		isRenderer := strings.Contains(line, "--type=renderer")
		pid, err1 := strconv.Atoi(parts[len(parts)-2])
		mem, err2 := strconv.ParseInt(parts[len(parts)-1], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}

		result = append(result, rendererProcessInfo{
			PID:        pid,
			MemMB:      mem / (1024 * 1024),
			Platform:   platform,
			ToolName:   displayName,
			IsRenderer: isRenderer,
		})
	}

	var renderers []rendererProcessInfo
	for _, p := range result {
		if p.IsRenderer {
			renderers = append(renderers, p)
		}
	}

	if len(renderers) > 0 {
		log.Printf("[memscan] Found %d renderer processes for %s", len(renderers), displayName)
	}

	return renderers
}

func (ms *MemoryScanner) FindTraeRendererPIDs() []rendererProcessInfo {
	return ms.findRendererPIDsForProcess("Trae CN.exe", "Trae CN", "trae")
}

func (ms *MemoryScanner) ScanForPattern(pid int, pattern string, contextBefore int, contextAfter int, maxResults int) []string {
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)
	if handle == 0 {
		return nil
	}
	defer procCloseHandle.Call(handle)

	patternBytes := []byte(pattern)
	var results []string
	var addr uintptr
	mbiSize := unsafe.Sizeof(memoryBasicInformation{})

	for {
		var mbi memoryBasicInformation
		ret, _, _ := procVirtualQueryEx.Call(
			handle,
			addr,
			uintptr(unsafe.Pointer(&mbi)),
			uintptr(mbiSize),
		)
		if ret == 0 {
			break
		}

		base := mbi.BaseAddress
		size := mbi.RegionSize

		if mbi.State == MEM_COMMIT && (mbi.Protect&READABLE) != 0 && size > 0 && size < 100*1024*1024 {
			buf := make([]byte, size)
			var bytesRead uintptr

			ret, _, _ = procReadProcessMem.Call(
				handle,
				base,
				uintptr(unsafe.Pointer(&buf[0])),
				size,
				uintptr(unsafe.Pointer(&bytesRead)),
			)
			if ret != 0 && bytesRead > 0 {
				data := buf[:bytesRead]
				idx := 0
				for {
					pos := ms.findBytes(data, patternBytes, idx)
					if pos == -1 {
						break
					}

					start := pos - contextBefore
					if start < 0 {
						start = 0
					}
					end := pos + contextAfter
					if end > len(data) {
						end = len(data)
					}

					text := ms.cleanPrintable(data[start:end])
					results = append(results, text)

					idx = pos + len(patternBytes)
					if len(results) >= maxResults {
						break
					}
				}
				if len(results) >= maxResults {
					break
				}
			}
		}

		nextAddr := base + size
		if nextAddr <= base {
			break
		}
		addr = nextAddr
	}

	return results
}

func (ms *MemoryScanner) ScanForPatternRaw(pid int, pattern string, contextBefore int, contextAfter int, maxResults int) []string {
	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)
	if handle == 0 {
		return nil
	}
	defer procCloseHandle.Call(handle)

	patternBytes := []byte(pattern)
	var results []string
	var addr uintptr
	mbiSize := unsafe.Sizeof(memoryBasicInformation{})

	for {
		var mbi memoryBasicInformation
		ret, _, _ := procVirtualQueryEx.Call(
			handle,
			addr,
			uintptr(unsafe.Pointer(&mbi)),
			uintptr(mbiSize),
		)
		if ret == 0 {
			break
		}

		base := mbi.BaseAddress
		size := mbi.RegionSize

		if mbi.State == MEM_COMMIT && (mbi.Protect&READABLE) != 0 && size > 0 && size < 100*1024*1024 {
			buf := make([]byte, size)
			var bytesRead uintptr

			ret, _, _ = procReadProcessMem.Call(
				handle,
				base,
				uintptr(unsafe.Pointer(&buf[0])),
				size,
				uintptr(unsafe.Pointer(&bytesRead)),
			)
			if ret != 0 && bytesRead > 0 {
				data := buf[:bytesRead]
				idx := 0
				for {
					pos := ms.findBytes(data, patternBytes, idx)
					if pos == -1 {
						break
					}

					start := pos - contextBefore
					if start < 0 {
						start = 0
					}
					end := pos + contextAfter
					if end > len(data) {
						end = len(data)
					}

					chunk := data[start:end]
					cleaned := make([]byte, 0, len(chunk))
					for _, b := range chunk {
						if b == 0 {
							continue
						}
						if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
							continue
						}
						cleaned = append(cleaned, b)
					}
					results = append(results, string(cleaned))

					idx = pos + len(patternBytes)
					if len(results) >= maxResults {
						break
					}
				}
				if len(results) >= maxResults {
					break
				}
			}
		}

		nextAddr := base + size
		if nextAddr <= base {
			break
		}
		addr = nextAddr
	}

	return results
}

func (ms *MemoryScanner) findBytes(data []byte, pattern []byte, startIdx int) int {
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx+len(pattern) > len(data) {
		return -1
	}
	for i := startIdx; i <= len(data)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if data[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func (ms *MemoryScanner) cleanPrintable(data []byte) string {
	var sb strings.Builder
	sb.Grow(len(data))
	i := 0
	for i < len(data) {
		b := data[i]
		if b >= 0x20 && b < 0x7f {
			sb.WriteByte(b)
			i++
		} else if b == '\n' || b == '\r' || b == '\t' {
			sb.WriteByte(b)
			i++
		} else if b == 0 {
			sb.WriteByte(' ')
			i++
		} else if b >= 0xC0 && b < 0xFE {
			var seqLen int
			if b < 0xE0 {
				seqLen = 2
			} else if b < 0xF0 {
				seqLen = 3
			} else {
				seqLen = 4
			}
			if i+seqLen <= len(data) && isValidUTF8Sequence(data[i:i+seqLen]) {
				sb.Write(data[i : i+seqLen])
				i += seqLen
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return sb.String()
}

func (ms *MemoryScanner) ExtractConversations() []AIConversationItem {
	renderers := ms.findAllRendererPIDs()
	if len(renderers) == 0 {
		log.Printf("[memscan] No IDE renderer processes found")
		return nil
	}

	log.Printf("[memscan] Found %d total renderer processes across all IDEs", len(renderers))

	var allItems []AIConversationItem

	for _, r := range renderers {
		if r.MemMB < 50 {
			log.Printf("[memscan] Skipping %s PID %d (%d MB, too small)", r.ToolName, r.PID, r.MemMB)
			continue
		}

		log.Printf("[memscan] Scanning %s PID %d (%d MB)", r.ToolName, r.PID, r.MemMB)

		items := ms.extractFromRenderer(r)
		for i := range items {
			items[i].Platform = r.Platform
		}
		allItems = append(allItems, items...)
	}

	allItems = ms.deduplicateItems(allItems)
	log.Printf("[memscan] Total unique conversation items: %d", len(allItems))
	return allItems
}

func (ms *MemoryScanner) extractFromRenderer(r rendererProcessInfo) []AIConversationItem {
	var allItems []AIConversationItem

	switch r.Platform {
	case "trae":
		allItems = append(allItems, ms.extractTraeConversations(r.PID)...)
	case "cursor":
		allItems = append(allItems, ms.extractCursorConversations(r.PID)...)
	case "codebuddy":
		allItems = append(allItems, ms.extractCodeBuddyConversations(r.PID)...)
	case "qoder":
		allItems = append(allItems, ms.extractQoderConversations(r.PID)...)
	default:
		allItems = append(allItems, ms.extractGenericConversations(r.PID, r.Platform)...)
	}

	return allItems
}

func (ms *MemoryScanner) extractTraeConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	sessionItems := ms.extractSessionList(pid)
	allItems = append(allItems, sessionItems...)

	responseItems := ms.extractAIResponses(pid)
	allItems = append(allItems, responseItems...)

	thoughtItems := ms.extractThoughtResponses(pid)
	allItems = append(allItems, thoughtItems...)

	inputItems := ms.extractUserInputs(pid)
	allItems = append(allItems, inputItems...)

	return allItems
}

func (ms *MemoryScanner) extractCursorConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	aiResponseItems := ms.extractCursorAIResponses(pid)
	allItems = append(allItems, aiResponseItems...)

	userInputItems := ms.extractCursorUserInputs(pid)
	allItems = append(allItems, userInputItems...)

	chatItems := ms.extractCursorChatData(pid)
	allItems = append(allItems, chatItems...)

	return allItems
}

func (ms *MemoryScanner) extractCodeBuddyConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	sessionItems := ms.extractSessionList(pid)
	allItems = append(allItems, sessionItems...)

	responseItems := ms.extractAIResponses(pid)
	allItems = append(allItems, responseItems...)

	thoughtItems := ms.extractThoughtResponses(pid)
	allItems = append(allItems, thoughtItems...)

	inputItems := ms.extractUserInputs(pid)
	allItems = append(allItems, inputItems...)

	return allItems
}

func (ms *MemoryScanner) extractQoderConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	sessionItems := ms.extractSessionList(pid)
	allItems = append(allItems, sessionItems...)

	responseItems := ms.extractAIResponses(pid)
	allItems = append(allItems, responseItems...)

	thoughtItems := ms.extractThoughtResponses(pid)
	allItems = append(allItems, thoughtItems...)

	inputItems := ms.extractUserInputs(pid)
	allItems = append(allItems, inputItems...)

	return allItems
}

func (ms *MemoryScanner) extractGenericConversations(pid int, platform string) []AIConversationItem {
	var allItems []AIConversationItem

	sessionItems := ms.extractSessionList(pid)
	allItems = append(allItems, sessionItems...)

	responseItems := ms.extractAIResponses(pid)
	allItems = append(allItems, responseItems...)

	thoughtItems := ms.extractThoughtResponses(pid)
	allItems = append(allItems, thoughtItems...)

	inputItems := ms.extractUserInputs(pid)
	allItems = append(allItems, inputItems...)

	return allItems
}

func (ms *MemoryScanner) extractSessionList(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, "session_status", 500, 500000, 20)
	if len(results) == 0 {
		return nil
	}

	var items []AIConversationItem
	sessionRe := regexp.MustCompile(`"session_id"\s*:\s*"([^"]+)"`)
	nameRe := regexp.MustCompile(`"name"\s*:\s*"([^"]*)"`)
	statusRe := regexp.MustCompile(`"session_status"\s*:\s*"([^"]*)"`)

	for _, text := range results {
		sids := sessionRe.FindAllStringSubmatch(text, -1)
		names := nameRe.FindAllStringSubmatch(text, -1)
		statuses := statusRe.FindAllStringSubmatch(text, -1)

		for i := range sids {
			if i >= len(names) || i >= len(statuses) {
				break
			}
			sid := sids[i][1]
			name := names[i][1]
			status := statuses[i][1]
			if name == "" {
				continue
			}
			items = append(items, AIConversationItem{
				Type:      "session",
				SessionID: sid,
				Name:      name,
				Status:    status,
			})
		}
	}
	return items
}

func (ms *MemoryScanner) extractAIResponses(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, "plan_item", 100, 500000, 200)
	if len(results) == 0 {
		return nil
	}

	thoughtRe := regexp.MustCompile(`"thought"\s*:\s*"((?:[^"\\]|\\.){20,}?)"`)
	sessionRe := regexp.MustCompile(`"session_id"\s*:\s*"([^"]*)"`)
	agentRe := regexp.MustCompile(`"agent_id"\s*:\s*"([^"]*)"`)

	seen := make(map[string]bool)
	var items []AIConversationItem
	for _, text := range results {
		sm := sessionRe.FindStringSubmatch(text)
		am := agentRe.FindStringSubmatch(text)
		matches := thoughtRe.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			thought := ms.unescapeJSON(m[1])
			if len(thought) < 20 {
				continue
			}
			if !ms.isReadableText(thought) {
				continue
			}
			key := thought
			if len(key) > 200 {
				key = key[:200]
			}
			if seen[key] {
				continue
			}
			seen[key] = true

			sessionID := ""
			agentID := ""
			if len(sm) > 1 {
				sessionID = sm[1]
			}
			if len(am) > 1 {
				agentID = am[1]
			}

			items = append(items, AIConversationItem{
				Type:      "ai_response",
				SessionID: sessionID,
				AgentID:   agentID,
				Thought:   thought,
			})
		}
	}
	return items
}

func (ms *MemoryScanner) extractThoughtResponses(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, `"thought":"`, 50, 500000, 200)
	if len(results) == 0 {
		return nil
	}

	thoughtRe := regexp.MustCompile(`"thought"\s*:\s*"((?:[^"\\]|\\.){20,}?)"`)
	sessionRe := regexp.MustCompile(`"session_id"\s*:\s*"([^"]*)"`)
	agentRe := regexp.MustCompile(`"agent_id"\s*:\s*"([^"]*)"`)

	seen := make(map[string]bool)
	var items []AIConversationItem
	for _, text := range results {
		sm := sessionRe.FindStringSubmatch(text)
		am := agentRe.FindStringSubmatch(text)
		matches := thoughtRe.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			thought := ms.unescapeJSON(m[1])
			if len(thought) < 20 {
				continue
			}
			if !ms.isReadableText(thought) {
				continue
			}
			key := thought
			if len(key) > 200 {
				key = key[:200]
			}
			if seen[key] {
				continue
			}
			seen[key] = true

			sessionID := ""
			agentID := ""
			if len(sm) > 1 {
				sessionID = sm[1]
			}
			if len(am) > 1 {
				agentID = am[1]
			}

			items = append(items, AIConversationItem{
				Type:      "ai_response",
				SessionID: sessionID,
				AgentID:   agentID,
				Thought:   thought,
			})
		}
	}
	return items
}

func (ms *MemoryScanner) isReadableText(s string) bool {
	if len(s) == 0 {
		return false
	}
	runes := []rune(s)
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
	if ratio < 0.8 {
		return false
	}
	if maxConsecutiveCJK >= 3 {
		return true
	}
	if cjkCount >= 5 {
		return true
	}
	v8HeapPatterns := []string{
		"MemoryScanner", "ScanForPattern", "cleanPrintable",
		"extractThought", "extractAIResponse", "regexp.MustCompile",
		"inherit", "prototype", "undefined", "constructor",
		"__proto__", "webpack", "chunkId", "moduleId",
	}
	for _, pat := range v8HeapPatterns {
		if strings.Contains(s, pat) {
			return false
		}
	}
	jsFileRe := regexp.MustCompile(`\b\w+\.js\b`)
	jsFileCount := len(jsFileRe.FindAllString(s, -1))
	if jsFileCount >= 3 {
		return false
	}
	goFileRe := regexp.MustCompile(`\b\w+\.go\b`)
	goFileCount := len(goFileRe.FindAllString(s, -1))
	if goFileCount >= 3 {
		return false
	}
	fileIdx := strings.Index(s, "file:///")
	if fileIdx >= 0 && fileIdx < 50 && cjkCount < 5 {
		return false
	}
	wordRe := regexp.MustCompile(`[a-zA-Z]{4,}`)
	words := wordRe.FindAllString(s, -1)
	longWordCount := 0
	for _, w := range words {
		if len(w) >= 6 {
			longWordCount++
		}
	}
	if longWordCount < 3 {
		return false
	}
	tsRe := regexp.MustCompile(`\.ts\b`)
	tsCount := len(tsRe.FindAllString(s, -1))
	if tsCount >= 3 {
		return false
	}
	spaceRe := regexp.MustCompile(` {3,}`)
	largeSpaces := len(spaceRe.FindAllString(s, -1))
	if largeSpaces > len(runes)/20 {
		return false
	}
	blankLines := strings.Count(s, "\n\n\n")
	if blankLines >= 2 {
		return false
	}
	nullLike := 0
	for _, r := range runes {
		if r == 0 || r == 0xFFFD || r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			nullLike++
		}
	}
	if nullLike > 0 {
		return false
	}
	return true
}

func (ms *MemoryScanner) unescapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\/`, "/")
	return s
}

func (ms *MemoryScanner) extractUserInputs(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, "inputText", 100, 200000, 50)
	if len(results) == 0 {
		return nil
	}

	var items []AIConversationItem
	inputRe := regexp.MustCompile(`"inputText"\s*:\s*"([^"]*)"`)

	for _, text := range results {
		matches := inputRe.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if len(m) > 1 && len(m[1]) >= 10 {
				items = append(items, AIConversationItem{
					Type:    "user_input",
					Content: m[1],
				})
			}
		}
	}
	return items
}

func (ms *MemoryScanner) extractCursorAIResponses(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	planItems := ms.ScanForPattern(pid, "plan_item", 100, 500000, 200)
	if len(planItems) > 0 {
		for _, text := range planItems {
			jsonObjects := ms.extractJSONObjects(text, 200)
			for _, obj := range jsonObjects {
				item := ms.parsePlanItemFromJSON(obj)
				if item != nil {
					allItems = append(allItems, *item)
				}
			}
		}
	}

	aiResultItems := ms.ScanForPattern(pid, `"role":"assistant"`, 100, 500000, 200)
	if len(aiResultItems) > 0 {
		for _, text := range aiResultItems {
			jsonObjects := ms.extractJSONObjects(text, 200)
			for _, obj := range jsonObjects {
				item := ms.parseAssistantMessage(obj)
				if item != nil {
					allItems = append(allItems, *item)
				}
			}
		}
	}

	composerItems := ms.ScanForPattern(pid, "ai-response", 100, 500000, 200)
	if len(composerItems) > 0 {
		for _, text := range composerItems {
			jsonObjects := ms.extractJSONObjects(text, 200)
			for _, obj := range jsonObjects {
				item := ms.parseAIResponseObject(obj)
				if item != nil {
					allItems = append(allItems, *item)
				}
			}
		}
	}

	return allItems
}

func (ms *MemoryScanner) extractCursorUserInputs(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	inputItems := ms.ScanForPattern(pid, "inputText", 100, 200000, 50)
	if len(inputItems) > 0 {
		inputRe := regexp.MustCompile(`"inputText"\s*:\s*"([^"]*)"`)
		for _, text := range inputItems {
			matches := inputRe.FindAllStringSubmatch(text, -1)
			for _, m := range matches {
				if len(m) > 1 && len(m[1]) >= 10 {
					allItems = append(allItems, AIConversationItem{
						Type:    "user_input",
						Content: m[1],
					})
				}
			}
		}
	}

	userMsgItems := ms.ScanForPattern(pid, `"role":"user"`, 100, 200000, 200)
	if len(userMsgItems) > 0 {
		for _, text := range userMsgItems {
			jsonObjects := ms.extractJSONObjects(text, 100)
			for _, obj := range jsonObjects {
				item := ms.parseUserMessage(obj)
				if item != nil {
					allItems = append(allItems, *item)
				}
			}
		}
	}

	return allItems
}

func (ms *MemoryScanner) extractCursorChatData(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, "composerData", 200, 500000, 200)
	if len(results) == 0 {
		return nil
	}

	var items []AIConversationItem
	for _, text := range results {
		jsonObjects := ms.extractJSONObjects(text, 300)
		for _, obj := range jsonObjects {
			item := ms.parseComposerData(obj)
			if item != nil {
				items = append(items, *item)
			}
		}
	}
	return items
}

func (ms *MemoryScanner) parseAssistantMessage(obj map[string]interface{}) *AIConversationItem {
	role, _ := obj["role"].(string)
	if role != "assistant" {
		return nil
	}

	content, _ := obj["content"].(string)
	if len(content) < 50 {
		return nil
	}

	return &AIConversationItem{
		Type:    "ai_response",
		Content: content,
	}
}

func (ms *MemoryScanner) parseUserMessage(obj map[string]interface{}) *AIConversationItem {
	role, _ := obj["role"].(string)
	if role != "user" {
		return nil
	}

	content, _ := obj["content"].(string)
	if len(content) < 10 {
		return nil
	}

	return &AIConversationItem{
		Type:    "user_input",
		Content: content,
	}
}

func (ms *MemoryScanner) parseAIResponseObject(obj map[string]interface{}) *AIConversationItem {
	typ, _ := obj["type"].(string)
	if typ != "ai-response" {
		return nil
	}

	data, _ := obj["data"].(map[string]interface{})
	if data == nil {
		return nil
	}

	text, _ := data["text"].(string)
	if len(text) < 50 {
		return nil
	}

	return &AIConversationItem{
		Type:    "ai_response",
		Content: text,
	}
}

func (ms *MemoryScanner) parseComposerData(obj map[string]interface{}) *AIConversationItem {
	composerData, _ := obj["composerData"].(map[string]interface{})
	if composerData == nil {
		return nil
	}

	conversation, _ := composerData["conversation"].(map[string]interface{})
	if conversation == nil {
		return nil
	}

	var items []AIConversationItem

	messages, _ := conversation["messages"].([]interface{})
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msgMap["role"].(string)
		text, _ := msgMap["text"].(string)

		switch role {
		case "assistant":
			if len(text) >= 50 {
				items = append(items, AIConversationItem{
					Type:    "ai_response",
					Content: text,
				})
			}
		case "user":
			if len(text) >= 10 {
				items = append(items, AIConversationItem{
					Type:    "user_input",
					Content: text,
				})
			}
		}
	}

	if len(items) > 0 {
		return &items[0]
	}
	return nil
}

func (ms *MemoryScanner) extractJSONObjects(text string, minSize int) []map[string]interface{} {
	var objects []map[string]interface{}
	i := 0
	for i < len(text) {
		if text[i] != '{' {
			i++
			continue
		}

		depth := 0
		start := i
		inString := false
		escape := false
		j := i

		for j < len(text) {
			c := text[j]
			if escape {
				escape = false
				j++
				continue
			}
			if c == '\\' && inString {
				escape = true
				j++
				continue
			}
			if c == '"' && !escape {
				inString = !inString
			} else if !inString {
				if c == '{' {
					depth++
				} else if c == '}' {
					depth--
					if depth == 0 {
						block := text[start : j+1]
						if len(block) >= minSize {
							var obj map[string]interface{}
							if err := json.Unmarshal([]byte(block), &obj); err == nil {
								objects = append(objects, obj)
							}
						}
						i = j + 1
						break
					}
				}
			}
			j++
		}
		if j >= len(text) {
			i++
		}
	}
	return objects
}

func (ms *MemoryScanner) parsePlanItemFromJSON(obj map[string]interface{}) *AIConversationItem {
	method, _ := obj["method"].(string)

	if strings.Contains(method, "stream") {
		params, _ := obj["params"].(map[string]interface{})
		data, _ := params["data"].(map[string]interface{})
		if data == nil {
			return nil
		}
		innerParams, _ := data["params"].(map[string]interface{})
		innerData, _ := innerParams["data"].(map[string]interface{})
		if innerData == nil {
			return nil
		}
		return ms.extractPlanItemFromData(innerData, data)
	}

	result, _ := obj["result"].(map[string]interface{})
	if result != nil {
		params, _ := result["params"].(map[string]interface{})
		if params == nil {
			return nil
		}
		data, _ := params["data"].(map[string]interface{})
		if data == nil {
			return nil
		}
		return ms.extractPlanItemFromData(data, data)
	}

	return nil
}

func (ms *MemoryScanner) extractPlanItemFromData(eventData map[string]interface{}, sessionSource map[string]interface{}) *AIConversationItem {
	event, _ := eventData["event"].(string)
	if event != "plan_item" {
		return nil
	}

	payload, _ := eventData["payload"].(map[string]interface{})
	if payload == nil {
		return nil
	}

	thought, _ := payload["thought"].(string)
	if len(thought) < 20 {
		return nil
	}

	agentID, _ := payload["agent_id"].(string)
	sessionID, _ := eventData["session_id"].(string)
	if sessionID == "" {
		sessionID, _ = sessionSource["session_id"].(string)
	}

	return &AIConversationItem{
		Type:      "ai_response",
		SessionID: sessionID,
		AgentID:   agentID,
		Thought:   thought,
	}
}

func (ms *MemoryScanner) deduplicateItems(items []AIConversationItem) []AIConversationItem {
	seen := make(map[string]bool)
	var result []AIConversationItem
	for _, item := range items {
		var key string
		if item.Thought != "" {
			key = fmt.Sprintf("thought:%s:%d", item.SessionID, len(item.Thought))
		} else if item.Content != "" {
			key = fmt.Sprintf("content:%s:%s", item.Platform, item.Content[:min(100, len(item.Content))])
		} else {
			key = fmt.Sprintf("session:%s:%s:%s", item.Platform, item.SessionID, item.Name)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}
