package main

import (
	"fmt"
	"os"
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

func main() {
	pid := 14108

	handle, _, _ := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0), uintptr(pid),
	)
	if handle == 0 {
		fmt.Printf("Cannot open PID %d\n", pid)
		os.Exit(1)
	}
	defer procCloseHandle.Call(handle)

	patterns := []string{"plan_item", "\"thought\"", "ai-response"}

	for _, pattern := range patterns {
		fmt.Printf("\n========== Pattern: %s ==========\n", pattern)
		patBytes := []byte(pattern)
		found := 0
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
					idx := 0
					for {
						pos := findBytes(data, patBytes, idx)
						if pos == -1 {
							break
						}

						if found < 3 {
							start := pos - 200
							if start < 0 {
								start = 0
							}
							end := pos + 1000
							if end > len(data) {
								end = len(data)
							}

							text := cleanPrintable(data[start:end])
							fmt.Printf("\n--- Match %d (offset %d) ---\n%s\n", found+1, pos, text)
						}
						found++
						idx = pos + len(patBytes)
					}
				}
			}

			nextAddr := base + size
			if nextAddr <= base {
				break
			}
			addr = nextAddr
		}
		fmt.Printf("\nTotal matches: %d\n", found)
	}
}

func findBytes(data []byte, pattern []byte, startIdx int) int {
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

func cleanPrintable(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		if b >= 0x20 && b < 0x7f {
			sb.WriteByte(b)
		} else if b == '\n' || b == '\r' || b == '\t' {
			sb.WriteByte(b)
		} else if b == 0 {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
