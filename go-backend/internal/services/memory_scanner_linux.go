//go:build linux

package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MemoryScanner struct{}

func NewMemoryScanner() *MemoryScanner {
	return &MemoryScanner{}
}

type linuxProcInfo struct {
	PID        int
	MemMB      int64
	Platform   string
	ToolName   string
	IsRenderer bool
}

func (ms *MemoryScanner) findAllRendererPIDs() []linuxProcInfo {
	var allRenderers []linuxProcInfo

	for _, tool := range supportedIDETolls {
		renderers := ms.findRendererPIDsForProcess(tool.ProcessName, tool.DisplayName, tool.Platform)
		allRenderers = append(allRenderers, renderers...)
	}

	return allRenderers
}

func (ms *MemoryScanner) findRendererPIDsForProcess(processName string, displayName string, platform string) []linuxProcInfo {
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		log.Printf("[memscan-linux] Cannot read /proc: %v", err)
		return nil
	}

	var result []linuxProcInfo
	linuxProcName := strings.TrimSuffix(processName, ".exe")

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		cmdline, err := os.ReadFile(filepath.Join(procDir, entry.Name(), "cmdline"))
		if err != nil {
			continue
		}

		cmdStr := string(cmdline)
		if !strings.Contains(cmdStr, linuxProcName) {
			continue
		}

		isRenderer := strings.Contains(cmdStr, "--type=renderer")

		stat, err := os.ReadFile(filepath.Join(procDir, entry.Name(), "stat"))
		if err != nil {
			continue
		}

		var memKB int64
		statParts := strings.Fields(string(stat))
		if len(statParts) >= 24 {
			memKB, _ = strconv.ParseInt(statParts[23], 10, 64)
		}

		result = append(result, linuxProcInfo{
			PID:        pid,
			MemMB:      memKB / 1024,
			Platform:   platform,
			ToolName:   displayName,
			IsRenderer: isRenderer,
		})
	}

	var renderers []linuxProcInfo
	for _, p := range result {
		if p.IsRenderer {
			renderers = append(renderers, p)
		}
	}

	if len(renderers) > 0 {
		log.Printf("[memscan-linux] Found %d renderer processes for %s", len(renderers), displayName)
	}

	return renderers
}

func (ms *MemoryScanner) FindTraeRendererPIDs() []linuxProcInfo {
	return ms.findRendererPIDsForProcess("Trae CN.exe", "Trae CN", "trae")
}

type memRegion struct {
	Start uintptr
	End   uintptr
}

func (ms *MemoryScanner) getReadableRegions(pid int) []memRegion {
	mapsPath := fmt.Sprintf("/proc/%d/maps", pid)
	data, err := os.ReadFile(mapsPath)
	if err != nil {
		return nil
	}

	var regions []memRegion
	readablePerm := regexp.MustCompile(`^[0-9a-f]+-[0-9a-f]+\s+[r-]`)

	for _, line := range strings.Split(string(data), "\n") {
		if !readablePerm.MatchString(line) {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 1 {
			continue
		}

		addrParts := strings.Split(parts[0], "-")
		if len(addrParts) != 2 {
			continue
		}

		start, err1 := strconv.ParseUint(addrParts[0], 16, 64)
		end, err2 := strconv.ParseUint(addrParts[1], 16, 64)
		if err1 != nil || err2 != nil {
			continue
		}

		size := end - start
		if size == 0 || size > 100*1024*1024 {
			continue
		}

		if strings.Contains(line, "[vsyscall]") ||
			strings.Contains(line, "[vvar]") ||
			strings.Contains(line, "[vdso]") {
			continue
		}

		regions = append(regions, memRegion{Start: uintptr(start), End: uintptr(end)})
	}

	return regions
}

func (ms *MemoryScanner) ScanForPattern(pid int, pattern string, contextBefore int, contextAfter int, maxResults int) []string {
	memPath := fmt.Sprintf("/proc/%d/mem", pid)
	f, err := os.Open(memPath)
	if err != nil {
		log.Printf("[memscan-linux] Cannot open %s: %v", memPath, err)
		return nil
	}
	defer f.Close()

	regions := ms.getReadableRegions(pid)
	if len(regions) == 0 {
		return nil
	}

	patternBytes := []byte(pattern)
	var results []string

	for _, region := range regions {
		size := region.End - region.Start
		if size == 0 || size > 100*1024*1024 {
			continue
		}

		buf := make([]byte, size)
		n, err := f.ReadAt(buf, int64(region.Start))
		if err != nil || n == 0 {
			continue
		}

		data := buf[:n]
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
		log.Printf("[memscan-linux] No IDE renderer processes found")
		return nil
	}

	log.Printf("[memscan-linux] Found %d total renderer processes across all IDEs", len(renderers))

	var allItems []AIConversationItem

	for _, r := range renderers {
		if r.MemMB < 50 {
			log.Printf("[memscan-linux] Skipping %s PID %d (%d MB, too small)", r.ToolName, r.PID, r.MemMB)
			continue
		}

		log.Printf("[memscan-linux] Scanning %s PID %d (%d MB)", r.ToolName, r.PID, r.MemMB)

		items := ms.extractFromRenderer(r)
		for i := range items {
			items[i].Platform = r.Platform
		}
		allItems = append(allItems, items...)
	}

	allItems = ms.deduplicateItems(allItems)
	log.Printf("[memscan-linux] Total unique conversation items: %d", len(allItems))
	return allItems
}

func (ms *MemoryScanner) extractFromRenderer(r linuxProcInfo) []AIConversationItem {
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
	allItems = append(allItems, ms.extractSessionList(pid)...)
	allItems = append(allItems, ms.extractAIResponses(pid)...)
	allItems = append(allItems, ms.extractThoughtResponses(pid)...)
	allItems = append(allItems, ms.extractUserInputs(pid)...)
	return allItems
}

func (ms *MemoryScanner) extractCursorConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem
	allItems = append(allItems, ms.extractCursorAIResponses(pid)...)
	allItems = append(allItems, ms.extractCursorUserInputs(pid)...)
	allItems = append(allItems, ms.extractCursorChatData(pid)...)
	return allItems
}

func (ms *MemoryScanner) extractCodeBuddyConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem
	allItems = append(allItems, ms.extractSessionList(pid)...)
	allItems = append(allItems, ms.extractAIResponses(pid)...)
	allItems = append(allItems, ms.extractThoughtResponses(pid)...)
	allItems = append(allItems, ms.extractUserInputs(pid)...)
	return allItems
}

func (ms *MemoryScanner) extractQoderConversations(pid int) []AIConversationItem {
	var allItems []AIConversationItem
	allItems = append(allItems, ms.extractSessionList(pid)...)
	allItems = append(allItems, ms.extractAIResponses(pid)...)
	allItems = append(allItems, ms.extractThoughtResponses(pid)...)
	allItems = append(allItems, ms.extractUserInputs(pid)...)
	return allItems
}

func (ms *MemoryScanner) extractGenericConversations(pid int, platform string) []AIConversationItem {
	var allItems []AIConversationItem
	allItems = append(allItems, ms.extractSessionList(pid)...)
	allItems = append(allItems, ms.extractAIResponses(pid)...)
	allItems = append(allItems, ms.extractThoughtResponses(pid)...)
	allItems = append(allItems, ms.extractUserInputs(pid)...)
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
	results := ms.ScanForPattern(pid, "plan_item", 100, 500000, 50)
	if len(results) == 0 {
		return nil
	}

	var items []AIConversationItem
	for _, text := range results {
		jsonObjects := ms.extractJSONObjects(text, 200)
		for _, obj := range jsonObjects {
			item := ms.parsePlanItemFromJSON(obj)
			if item != nil {
				items = append(items, *item)
			}
		}
	}
	return items
}

func (ms *MemoryScanner) extractThoughtResponses(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, `"thought"`, 50, 500000, 50)
	if len(results) == 0 {
		return nil
	}

	var items []AIConversationItem
	for _, text := range results {
		jsonObjects := ms.extractJSONObjects(text, 100)
		for _, obj := range jsonObjects {
			item := ms.parseThoughtObject(obj)
			if item != nil {
				items = append(items, *item)
			}
		}
	}
	return items
}

func (ms *MemoryScanner) parseThoughtObject(obj map[string]interface{}) *AIConversationItem {
	thought, _ := obj["thought"].(string)
	if len(thought) < 50 {
		return nil
	}

	sessionID, _ := obj["session_id"].(string)
	agentID, _ := obj["agent_id"].(string)

	return &AIConversationItem{
		Type:      "ai_response",
		SessionID: sessionID,
		AgentID:   agentID,
		Thought:   thought,
	}
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

	planItems := ms.ScanForPattern(pid, "plan_item", 100, 500000, 50)
	for _, text := range planItems {
		for _, obj := range ms.extractJSONObjects(text, 200) {
			if item := ms.parsePlanItemFromJSON(obj); item != nil {
				allItems = append(allItems, *item)
			}
		}
	}

	aiResultItems := ms.ScanForPattern(pid, `"role":"assistant"`, 100, 500000, 30)
	for _, text := range aiResultItems {
		for _, obj := range ms.extractJSONObjects(text, 200) {
			if item := ms.parseAssistantMessage(obj); item != nil {
				allItems = append(allItems, *item)
			}
		}
	}

	composerItems := ms.ScanForPattern(pid, "ai-response", 100, 500000, 30)
	for _, text := range composerItems {
		for _, obj := range ms.extractJSONObjects(text, 200) {
			if item := ms.parseAIResponseObject(obj); item != nil {
				allItems = append(allItems, *item)
			}
		}
	}

	return allItems
}

func (ms *MemoryScanner) extractCursorUserInputs(pid int) []AIConversationItem {
	var allItems []AIConversationItem

	inputItems := ms.ScanForPattern(pid, "inputText", 100, 200000, 50)
	inputRe := regexp.MustCompile(`"inputText"\s*:\s*"([^"]*)"`)
	for _, text := range inputItems {
		for _, m := range inputRe.FindAllStringSubmatch(text, -1) {
			if len(m) > 1 && len(m[1]) >= 10 {
				allItems = append(allItems, AIConversationItem{Type: "user_input", Content: m[1]})
			}
		}
	}

	userMsgItems := ms.ScanForPattern(pid, `"role":"user"`, 100, 200000, 30)
	for _, text := range userMsgItems {
		for _, obj := range ms.extractJSONObjects(text, 100) {
			if item := ms.parseUserMessage(obj); item != nil {
				allItems = append(allItems, *item)
			}
		}
	}

	return allItems
}

func (ms *MemoryScanner) extractCursorChatData(pid int) []AIConversationItem {
	results := ms.ScanForPattern(pid, "composerData", 200, 500000, 30)
	if len(results) == 0 {
		return nil
	}

	var items []AIConversationItem
	for _, text := range results {
		for _, obj := range ms.extractJSONObjects(text, 300) {
			if item := ms.parseComposerData(obj); item != nil {
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
	return &AIConversationItem{Type: "ai_response", Content: content}
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
	return &AIConversationItem{Type: "user_input", Content: content}
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
	return &AIConversationItem{Type: "ai_response", Content: text}
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
				items = append(items, AIConversationItem{Type: "ai_response", Content: text})
			}
		case "user":
			if len(text) >= 10 {
				items = append(items, AIConversationItem{Type: "user_input", Content: text})
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
	if len(thought) < 50 {
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

func (ms *MemoryScanner) findProcessByName(name string) []int {
	var pids []int
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil {
			continue
		}
		if strings.Contains(string(cmdline), name) {
			pids = append(pids, pid)
		}
	}
	return pids
}

func init() {
	_ = exec.Command
	_ = regexp.MustCompile
}
