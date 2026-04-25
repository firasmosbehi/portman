# Agent Context: portman

## Project Overview
Build `portman`, a clean CLI for managing local ports and processes. List listening ports with owning processes. Kill processes by port. Check if ports are free. Suggest next available port. Essential for developers running multiple microservices locally who are tired of "Port 3000 is already in use" errors.

## Core Philosophy
- **Cross-platform**: One command that works identically on macOS, Linux, and Windows.
- **Clean output**: Colorized, sortable tables. No parsing `lsof` or `netstat` yourself.
- **Safe defaults**: Always confirm before killing. Never kill without explicit permission.
- **Project-aware**: `portman.yml` declares expected ports for health checks.

## Tech Stack
- **Primary**: Go (single binary, cross-platform process and network APIs).
- **Process info**: Platform-specific implementations using `lsof` (Unix), `netstat` (Windows), or native APIs.

## Commands & Features

### Core Commands
| Command | Description |
|---------|-------------|
| `portman list` | List all listening ports with process info. |
| `portman list --port 3000` | Show only port 3000. |
| `portman check <port>` | Report if port is free or in use. |
| `portman kill <port>` | Find and kill process using port (with confirmation). |
| `portman next` | Suggest the next available port in a range. |
| `portman watch <port>` | Monitor a port and notify when it becomes available. |
| `portman status` | Check project services against `portman.yml` registry. |

### List Output
```
$ portman list

Listening Ports

PORT    PROTOCOL  PROCESS        PID    USER       AGE
─────────────────────────────────────────────────────────
3000    tcp       node           4521   alice      2h
3001    tcp       node           4522   alice      2h
5432    tcp       postgres       1204   postgres   3d
6379    tcp       redis-server   1198   redis      3d
8080    tcp       python         8910   bob        15m

? Kill process on port 8080? (y/N)
```

### Project Registry (`portman.yml`)
```yaml
services:
  - name: web
    port: 3000
    command: npm run dev

  - name: api
    port: 3001
    command: npm run api

  - name: db
    port: 5432
    health_check: pg_isready

  - name: cache
    port: 6379
```

### Status Output
```
$ portman status

Project Services

SERVICE  EXPECTED  ACTUAL  STATUS
───────────────────────────────────
web      3000      3000    ✓ running
api      3001      3001    ✓ running
db       5432      5432    ✓ healthy
cache    6379      6379    ✓ running

All services healthy.
```

### Key Differentiators
1. **Cross-Platform Uniformity**: Same commands, same output on all OSes. No need to remember `lsof -i :3000` on macOS vs `netstat -ano | findstr :3000` on Windows.
2. **Watch Mode**: `portman watch 3000` polls every second and prints "Port 3000 is now available" when the process exits. Useful for waiting on slow startups.
3. **Project Health Check**: `portman status` validates the entire local stack. Know instantly if a service died.
4. **Next Port Suggestion**: `portman next --range 3000-3010` scans the range and returns the first available port. Eliminates guesswork.

## Architecture
```
portman/
├── cmd/
│   ├── list.go
│   ├── check.go
│   ├── kill.go
│   ├── next.go
│   ├── watch.go
│   └── status.go
├── internal/
│   ├── platform/    # OS-specific process/port resolution
│   │   ├── darwin.go
│   │   ├── linux.go
│   │   └── windows.go
│   ├── scanner/     # Port scanning and process lookup
│   ├── registry/    # portman.yml parsing
│   ├── health/      # Health check execution
│   └── reporter/    # Terminal output formatting
├── pkg/
│   └── models/
└── main.go
```

## Implementation Notes
- **Unix (macOS/Linux)**: Parse `lsof -i -P -n -F` output (machine-readable format) or `ss -tlnp` on Linux. Extract port, PID, process name.
- **Windows**: Use `netstat -ano` for ports, then `tasklist /FI "PID eq <pid>"` for process names. Or use Windows APIs via `golang.org/x/sys/windows`.
- **Process Killing**: Use `os.FindProcess(pid).Kill()` (cross-platform in Go) or shell out to `kill`/`taskkill`.
- **Watch Mode**: Simple polling loop with configurable interval. Use ANSI escape sequences to update a single line in terminal.
- **Health Checks**: Execute configured command and check exit code. Default to TCP port open check if no command specified.

## Testing Strategy
- **Unit tests**: Mock platform-specific command outputs. Test parsing logic.
- **Integration tests**: Start real processes on known ports, test list/check/kill operations.
- **Cross-platform tests**: Run CI on macOS, Ubuntu, and Windows runners.

## Distribution
- GoReleaser + Homebrew.
- Standalone binary.

## Success Metrics
- List all listening ports in < 1 second.
- Kill process by port in < 3 seconds with confirmation.
- Port availability check returns accurate result 100% of the time.
