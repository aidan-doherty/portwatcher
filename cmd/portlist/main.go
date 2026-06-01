package main

import (
	"fmt"

	"ad/portwatcher/internals/ports"
)

func main() {
	fmt.Println("Port List:")
	portsList, err := ports.GetPortList()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	// collect unique PIDs
	pidSet := make(map[uint32]struct{})
	var pids []uint32
	for _, port := range portsList {
		if port.PID == 0 {
			continue
		}
		if _, ok := pidSet[port.PID]; !ok {
			pidSet[port.PID] = struct{}{}
			pids = append(pids, port.PID)
		}
	}

	// resolve process names
	names := map[uint32]string{}
	if len(pids) > 0 {
		if m, err := ports.ProcessNames(pids); err == nil {
			names = m
		}
	}

	for _, port := range portsList {
		procName := names[port.PID]
		if procName != "" {
			fmt.Printf("Protocol: %s, Local IP: %s, Local Port: %d, PID: %d (%s), State: %s\n", port.Protocol, port.LocalIP, port.LocalPort, port.PID, procName, port.State)
		} else {
			fmt.Printf("Protocol: %s, Local IP: %s, Local Port: %d, PID: %d, State: %s\n", port.Protocol, port.LocalIP, port.LocalPort, port.PID, port.State)
		}
	}
}
