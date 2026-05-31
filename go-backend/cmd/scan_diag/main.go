package main

import (
	"fmt"
	"os"
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

func countPattern(handle uintptr, pattern string) int {
	count := 0
	mbiSize := unsafe.Sizeof(memoryBasicInformation{})
	var addr uintptr
	patBytes := []byte(pattern)

	for {
		var mbi memoryBasicInformation
		ret, _, _ := procVirtualQueryEx.Call(
			handle, addr,
			uintptr(unsafe.Pointer(&mbi)), uintptr(mbiSize),
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
				handle, base,
				uintptr(unsafe.Pointer(&buf[0])), size,
				uintptr(unsafe.Pointer(&bytesRead)),
			)
			if ret != 0 && bytesRead > 0 {
				data := buf[:bytesRead]
				count += countBytes(data, patBytes)
			}
		}

		nextAddr := base + size
		if nextAddr <= base {
			break
		}
		addr = nextAddr
	}
	return count
}

func countBytes(data []byte, pattern []byte) int {
	count := 0
	idx := 0
	for {
		pos := findBytes(data, pattern, idx)
		if pos == -1 {
			break
		}
		count++
		idx = pos + len(pattern)
	}
	return count
}

func findBytes(data []byte, pattern []byte, startIdx int) int {
	if startIdx < 0 {
		startIdx = 0
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

func getMemStats(handle uintptr) (int, int, uint64) {
	totalRegions := 0
	readableRegions := 0
	var totalBytes uint64
	mbiSize := unsafe.Sizeof(memoryBasicInformation{})
	var addr uintptr

	for {
		var mbi memoryBasicInformation
		ret, _, _ := procVirtualQueryEx.Call(
			handle, addr,
			uintptr(unsafe.Pointer(&mbi)), uintptr(mbiSize),
		)
		if ret == 0 {
			break
		}
		totalRegions++
		base := mbi.BaseAddress
		size := mbi.RegionSize

		if mbi.State == MEM_COMMIT && (mbi.Protect&READABLE) != 0 && size > 0 && size < 100*1024*1024 {
			readableRegions++
			totalBytes += uint64(size)
		}

		nextAddr := base + size
		if nextAddr <= base {
			break
		}
		addr = nextAddr
	}
	return totalRegions, readableRegions, totalBytes
}

func main() {
	pids := []int{13444, 14108, 47124}

	patterns := []struct {
		name    string
		pattern string
	}{
		{"plan_item", "plan_item"},
		{"session_status", "session_status"},
		{"inputText", "inputText"},
		{"role:assistant", "\"role\":\"assistant\""},
		{"role:user", "\"role\":\"user\""},
		{"thought", "\"thought\""},
		{"content:", "\"content\":"},
		{"ai-response", "ai-response"},
		{"composerData", "composerData"},
		{"type:assistant", "\"type\":\"assistant\""},
		{"type:user", "\"type\":\"user\""},
		{"agent_id", "\"agent_id\""},
		{"session_id", "\"session_id\""},
		{"stream", "\"stream\""},
	}

	for _, pid := range pids {
		handle, _, _ := procOpenProcess.Call(
			uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
			uintptr(0), uintptr(pid),
		)
		if handle == 0 {
			fmt.Printf("\n=== PID %d: CANNOT OPEN (error=%d) ===\n", pid, syscall.GetLastError())
			continue
		}

		totalR, readR, totalB := getMemStats(handle)
		fmt.Printf("\n=== PID %d: %d regions, %d readable, %d MB ===\n", pid, totalR, readR, totalB/1024/1024)

		for _, p := range patterns {
			count := countPattern(handle, p.pattern)
			if count > 0 {
				fmt.Printf("  %-20s : %d\n", p.name, count)
			}
		}

		procCloseHandle.Call(handle)
	}

	fmt.Println("\nDone.")
	os.Exit(0)
}
