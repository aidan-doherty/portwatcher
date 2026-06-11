# portlist

![Main view](screenshots/main.jpg)

A TUI tool to list active network ports and their owning processes on Windows. Built with Go and [tview](https://github.com/rivo/tview).

## Controls

| Key | Action |
|-----|--------|
| `/` | Open search (filter by port or process name) |
| `Esc` | Close search / clear filter |
| `n` | Sort by process name |
| `p` | Sort by port number |
| `t` | Sort by protocol |
| `o` | Toggle ascending/descending order |
| `r` | Refresh |
| `q` | Quit |

## Usage

```
portlist.exe
```

By default `portlist` uses the Win32 IP Helper API for fast port enumeration. If you encounter any issues, fall back to the `netstat` method:

```
portlist.exe --netstat
```

Press `/` to start typing a search query — results filter in real-time as you type. Press `Esc` to clear the search and return to the full list.

![Search active](screenshots/search.jpg)

## Benchmarks

The repo includes benchmarks comparing the `netstat`-based implementation (`GetPortListNetstat`) against the Win32 IP Helper API (`GetPortList`).

Run them on Windows:

```
go test -bench=. -benchtime=3s ./internals/ports/
```
