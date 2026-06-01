package ports

import (
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessName returns the executable (image) name for the given PID on Windows.
// If the PID is not found it returns an empty string and no error.
func ProcessName(pid uint32) (string, error) {
	// Use tasklist to get a friendly image name for the PID.
	// Example output (CSV):
	// "Image Name","PID","Session Name","Session#","Mem Usage"
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(int(pid)), "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tasklist failed: %w", err)
	}

	s := strings.TrimSpace(string(out))
	if s == "" {
		return "", nil
	}

	// tasklist may return an informational line when no tasks match
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
	// trim surrounding quotes if any (csv.Reader already does this)
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
