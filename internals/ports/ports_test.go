//go:build windows

package ports

import "testing"

var benchResult []PortInfo

func BenchmarkGetPortListNetstat(b *testing.B) {
	var r []PortInfo
	for i := 0; i < b.N; i++ {
		r, _ = GetPortListNetstat()
	}
	benchResult = r
}

func BenchmarkGetPortList(b *testing.B) {
	var r []PortInfo
	for i := 0; i < b.N; i++ {
		r, _ = GetPortList()
	}
	benchResult = r
}
