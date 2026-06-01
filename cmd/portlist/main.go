package main

import (
	"sort"
	"strconv"
	"time"

	"ad/portwatcher/internals/ports"
	proc "ad/portwatcher/internals/process"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SortField int

const (
	SortByPort SortField = iota
	SortByProcess
	SortByProtocol
)

func main() {
	app := tview.NewApplication()

	header := tview.NewTextView().SetDynamicColors(true)
	header.SetBorder(true).SetTitle("PortWatcher — Controls: [yellow]n[white]=name [yellow]p[white]=port [yellow]t[white]=protocol [yellow]r[white]=refresh [yellow]o[white]=order [yellow]q[white]=quit")

	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true).SetTitle("Ports")

	footer := tview.NewTextView().SetDynamicColors(true)
	footer.SetBorder(true).SetTitle("Status")

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(table, 0, 1, true).
		AddItem(footer, 3, 0, false)

	sortField := SortByPort
	asc := true

	loadAndRender := func() {
		table.Clear()
		portsList, err := ports.GetPortList()
		if err != nil {
			footer.SetText("Error: " + err.Error())
			return
		}

		pidSet := make(map[uint32]struct{})
		var pids []uint32
		for _, p := range portsList {
			if p.PID == 0 {
				continue
			}
			if _, ok := pidSet[p.PID]; !ok {
				pidSet[p.PID] = struct{}{}
				pids = append(pids, p.PID)
			}
		}

		names, _ := proc.ProcessNames(pids)

		type row struct {
			proto string
			ip    string
			port  uint16
			pid   uint32
			name  string
			state string
		}
		var rows []row
		for _, p := range portsList {
			rows = append(rows, row{proto: p.Protocol, ip: p.LocalIP, port: p.LocalPort, pid: p.PID, name: names[p.PID], state: p.State})
		}

		cmp := func(i, j int) bool { return true }
		switch sortField {
		case SortByPort:
			if asc {
				cmp = func(i, j int) bool { return rows[i].port < rows[j].port }
			} else {
				cmp = func(i, j int) bool { return rows[i].port > rows[j].port }
			}
		case SortByProcess:
			if asc {
				cmp = func(i, j int) bool { return rows[i].name < rows[j].name }
			} else {
				cmp = func(i, j int) bool { return rows[i].name > rows[j].name }
			}
		case SortByProtocol:
			if asc {
				cmp = func(i, j int) bool { return rows[i].proto < rows[j].proto }
			} else {
				cmp = func(i, j int) bool { return rows[i].proto > rows[j].proto }
			}
		}
		sort.Slice(rows, cmp)

		headers := []string{"Protocol", "Local IP", "Port", "PID", "Process", "State"}
		for c, h := range headers {
			table.SetCell(0, c, tview.NewTableCell("[::b]"+h).SetSelectable(false))
		}
		for r, rw := range rows {
			table.SetCell(r+1, 0, tview.NewTableCell(rw.proto))
			table.SetCell(r+1, 1, tview.NewTableCell(rw.ip))
			table.SetCell(r+1, 2, tview.NewTableCell(strconv.Itoa(int(rw.port))))
			pidStr := ""
			if rw.pid != 0 {
				pidStr = strconv.Itoa(int(rw.pid))
			}
			table.SetCell(r+1, 3, tview.NewTableCell(pidStr))
			table.SetCell(r+1, 4, tview.NewTableCell(rw.name))
			table.SetCell(r+1, 5, tview.NewTableCell(rw.state))
		}

		header.SetText("Sort: " + func() string {
			s := ""
			switch sortField {
			case SortByPort:
				s = "Port"
			case SortByProcess:
				s = "Process"
			case SortByProtocol:
				s = "Protocol"
			}
			if asc {
				s += " (asc)"
			} else {
				s += " (desc)"
			}
			return s
		}())

		footer.SetText("Last refresh: " + time.Now().Format(time.RFC1123))
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			app.Stop()
		case 'n', 'N':
			sortField = SortByProcess
			loadAndRender()
		case 'p', 'P':
			sortField = SortByPort
			loadAndRender()
		case 't', 'T':
			sortField = SortByProtocol
			loadAndRender()
		case 'o', 'O':
			asc = !asc
			loadAndRender()
		case 'r', 'R':
			loadAndRender()
		}
		return event
	})

	loadAndRender()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
