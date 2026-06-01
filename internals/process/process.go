package process

import (
	"encoding/csv"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procCloseHandle                = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	maxPath                           = 260
)

// ProcessName returns the executable (image) name for the given PID on Windows.
// It prefers the native Windows API (`QueryFullProcessImageNameW`) and falls
// back to `tasklist` CSV parsing when the API is unavailable or fails.
func ProcessName(pid uint32) (string, error) {
	if pid == 0 {
		return "", nil
	}

	// Try native API first
	h, _, err := procOpenProcess.Call(uintptr(PROCESS_QUERY_LIMITED_INFORMATION), uintptr(0), uintptr(pid))
	if h != 0 {
		defer procCloseHandle.Call(h)

		buf := make([]uint16, maxPath)
		size := uint32(len(buf))
		ret, _, _ := procQueryFullProcessImageNameW.Call(h, uintptr(0), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
		if ret != 0 {
			full := syscall.UTF16ToString(buf[:size])
			return filepath.Base(full), nil
		}
		// fall through to tasklist fallback
	}

	// Fallback: tasklist CSV parsing
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(int(pid)), "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		// return empty without error when process isn't accessible
		return "", nil
	}

	s := strings.TrimSpace(string(out))
	if s == "" {
		return "", nil
	}
	if strings.HasPrefix(s, "INFO:") || strings.Contains(strings.ToLower(s), "no tasks are running") {
		return "", nil
	}
	r := csv.NewReader(strings.NewReader(s))
	r.FieldsPerRecord = -1
	recs, err := r.ReadAll()
	if err != nil || len(recs) == 0 || len(recs[0]) == 0 {
		return "", nil
	}
	name := strings.TrimSpace(recs[0][0])
	name = strings.Trim(name, "\"\n\r ")
	return name, nil
}

// ProcessNames performs batch lookup for multiple PIDs. Missing PIDs will
// be omitted from the returned map.
func ProcessNames(pids []uint32) (map[uint32]string, error) {
	m := make(map[uint32]string, len(pids))
	for _, pid := range pids {
		name, err := ProcessName(pid)
		if err != nil {
			return nil, err
		}
		if name != "" {
			m[pid] = name
		}
	}
	return m, nil
}
