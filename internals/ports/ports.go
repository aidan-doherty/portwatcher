package ports

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type PortInfo struct {
	Protocol  string
	LocalIP   string
	LocalPort uint16
	PID       uint32
	State     string
}

// GetPortListNetstat runs `netstat -ano` and parses TCP/UDP listeners/connections.
func GetPortListNetstat() ([]PortInfo, error) {
	cmd := exec.Command("netstat", "-ano")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ports []PortInfo
	scanner := bufio.NewScanner(bytes.NewReader(out))
	started := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// header line indicates following entries
		if strings.HasPrefix(line, "Proto") {
			started = true
			continue
		}
		if !started {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		proto := strings.ToUpper(fields[0])
		local := fields[1]
		pidStr := fields[len(fields)-1]
		state := ""
		if proto == "TCP" {
			if len(fields) >= 5 {
				state = fields[3]
			}
		}

		// split local into ip and port (take last ':' to handle IPv6)
		ip := local
		portNum := uint16(0)
		if idx := strings.LastIndex(local, ":"); idx != -1 {
			ip = local[:idx]
			pstr := local[idx+1:]
			if p, err := strconv.Atoi(pstr); err == nil {
				portNum = uint16(p)
			}
		}

		pid := uint32(0)
		if p, err := strconv.Atoi(pidStr); err == nil {
			pid = uint32(p)
		}

		ports = append(ports, PortInfo{
			Protocol:  strings.ToLower(proto),
			LocalIP:   ip,
			LocalPort: portNum,
			PID:       pid,
			State:     state,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ports, nil
}
