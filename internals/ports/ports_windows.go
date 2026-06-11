//go:build windows

package ports

import (
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var iphlpapi = syscall.NewLazyDLL("iphlpapi.dll")

var (
	procGetExtendedTcpTable = iphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = iphlpapi.NewProc("GetExtendedUdpTable")
)

type _MIB_TCPROW_OWNER_PID struct {
	DwState      uint32
	DwLocalAddr  [4]byte
	DwLocalPort  uint32
	DwRemoteAddr [4]byte
	DwRemotePort uint32
	DwOwningPid  uint32
}

type _MIB_TCPTABLE_OWNER_PID struct {
	DwNumEntries uint32
	Table        [1]_MIB_TCPROW_OWNER_PID
}

type _MIB_TCP6ROW_OWNER_PID struct {
	UcLocalAddr    [16]byte
	DwLocalScopeId uint32
	DwLocalPort    uint32
	UcRemoteAddr   [16]byte
	DwRemoteScopeId uint32
	DwRemotePort   uint32
	DwState        uint32
	DwOwningPid    uint32
}

type _MIB_TCP6TABLE_OWNER_PID struct {
	DwNumEntries uint32
	Table        [1]_MIB_TCP6ROW_OWNER_PID
}

type _MIB_UDPROW_OWNER_PID struct {
	DwLocalAddr [4]byte
	DwLocalPort uint32
	DwOwningPid uint32
}

type _MIB_UDPTABLE_OWNER_PID struct {
	DwNumEntries uint32
	Table        [1]_MIB_UDPROW_OWNER_PID
}

type _MIB_UDP6ROW_OWNER_PID struct {
	UcLocalAddr    [16]byte
	DwLocalScopeId uint32
	DwLocalPort    uint32
	DwOwningPid    uint32
}

type _MIB_UDP6TABLE_OWNER_PID struct {
	DwNumEntries uint32
	Table        [1]_MIB_UDP6ROW_OWNER_PID
}

const (
	_AF_INET  = 2
	_AF_INET6 = 23

	_TCP_TABLE_OWNER_PID_ALL = 5
	_UDP_TABLE_OWNER_PID     = 1

	_ERROR_INSUFFICIENT_BUFFER = 122
)

var tcpState = map[uint32]string{
	1:  "CLOSED",
	2:  "LISTENING",
	3:  "SYN_SENT",
	4:  "SYN_RCVD",
	5:  "ESTABLISHED",
	6:  "FIN_WAIT1",
	7:  "FIN_WAIT2",
	8:  "CLOSE_WAIT",
	9:  "CLOSING",
	10: "LAST_ACK",
	11: "TIME_WAIT",
	12: "DELETE_TCB",
}

func GetPortList() ([]PortInfo, error) {
	var ports []PortInfo

	tcp4, err := getTCP4()
	if err != nil {
		return nil, err
	}
	ports = append(ports, tcp4...)

	tcp6, err := getTCP6()
	if err != nil {
		return nil, err
	}
	ports = append(ports, tcp6...)

	udp4, err := getUDP4()
	if err != nil {
		return nil, err
	}
	ports = append(ports, udp4...)

	udp6, err := getUDP6()
	if err != nil {
		return nil, err
	}
	ports = append(ports, udp6...)

	return ports, nil
}

func getTCP4() ([]PortInfo, error) {
	buf, err := getTable(procGetExtendedTcpTable.Addr(), _AF_INET, _TCP_TABLE_OWNER_PID_ALL)
	if err != nil {
		return nil, err
	}
	table := (*_MIB_TCPTABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	n := table.DwNumEntries
	rows := unsafe.Slice(&table.Table[0], n)

	ports := make([]PortInfo, 0, n)
	for _, r := range rows {
		ip := net.IPv4(r.DwLocalAddr[0], r.DwLocalAddr[1], r.DwLocalAddr[2], r.DwLocalAddr[3]).String()
		port := windows.Ntohs(uint16(r.DwLocalPort))
		ports = append(ports, PortInfo{
			Protocol:  "tcp",
			LocalIP:   ip,
			LocalPort: port,
			PID:       r.DwOwningPid,
			State:     tcpState[r.DwState],
		})
	}
	return ports, nil
}

func getTCP6() ([]PortInfo, error) {
	buf, err := getTable(procGetExtendedTcpTable.Addr(), _AF_INET6, _TCP_TABLE_OWNER_PID_ALL)
	if err != nil {
		return nil, err
	}
	table := (*_MIB_TCP6TABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	n := table.DwNumEntries
	rows := unsafe.Slice(&table.Table[0], n)

	ports := make([]PortInfo, 0, n)
	for _, r := range rows {
		ip := formatIPv6(r.UcLocalAddr)
		port := windows.Ntohs(uint16(r.DwLocalPort))
		ports = append(ports, PortInfo{
			Protocol:  "tcp",
			LocalIP:   ip,
			LocalPort: port,
			PID:       r.DwOwningPid,
			State:     tcpState[r.DwState],
		})
	}
	return ports, nil
}

func getUDP4() ([]PortInfo, error) {
	buf, err := getTable(procGetExtendedUdpTable.Addr(), _AF_INET, _UDP_TABLE_OWNER_PID)
	if err != nil {
		return nil, err
	}
	table := (*_MIB_UDPTABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	n := table.DwNumEntries
	rows := unsafe.Slice(&table.Table[0], n)

	ports := make([]PortInfo, 0, n)
	for _, r := range rows {
		ip := net.IPv4(r.DwLocalAddr[0], r.DwLocalAddr[1], r.DwLocalAddr[2], r.DwLocalAddr[3]).String()
		port := windows.Ntohs(uint16(r.DwLocalPort))
		ports = append(ports, PortInfo{
			Protocol:  "udp",
			LocalIP:   ip,
			LocalPort: port,
			PID:       r.DwOwningPid,
			State:     "",
		})
	}
	return ports, nil
}

func getUDP6() ([]PortInfo, error) {
	buf, err := getTable(procGetExtendedUdpTable.Addr(), _AF_INET6, _UDP_TABLE_OWNER_PID)
	if err != nil {
		return nil, err
	}
	table := (*_MIB_UDP6TABLE_OWNER_PID)(unsafe.Pointer(&buf[0]))
	n := table.DwNumEntries
	rows := unsafe.Slice(&table.Table[0], n)

	ports := make([]PortInfo, 0, n)
	for _, r := range rows {
		ip := formatIPv6(r.UcLocalAddr)
		port := windows.Ntohs(uint16(r.DwLocalPort))
		ports = append(ports, PortInfo{
			Protocol:  "udp",
			LocalIP:   ip,
			LocalPort: port,
			PID:       r.DwOwningPid,
			State:     "",
		})
	}
	return ports, nil
}

func formatIPv6(addr [16]byte) string {
	return "[" + net.IP(addr[:]).String() + "]"
}

func getTable(proc uintptr, af, class uint32) ([]byte, error) {
	bufSize := uint32(64 * 1024)
	buf := make([]byte, bufSize)
	for i := 0; i < 3; i++ {
		r1, _, _ := syscall.Syscall6(proc, 6,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&bufSize)),
			1,
			uintptr(af),
			uintptr(class),
			0,
		)
		if r1 == 0 {
			return buf[:bufSize], nil
		}
		if r1 != _ERROR_INSUFFICIENT_BUFFER {
			return nil, syscall.Errno(r1)
		}
		buf = make([]byte, bufSize)
	}
	return nil, syscall.Errno(_ERROR_INSUFFICIENT_BUFFER)
}
