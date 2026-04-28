# Agent Context: portman

## Project Overview

`portman` is a cross-platform CLI tool written in Go for managing local ports and processes. It lets developers list listening ports, check availability, kill processes by port, find the next free port in a range, watch ports until they become available, and validate project services against a `portman.yml` registry.

The project is a single Go binary with no runtime dependencies beyond platform-native tools (`lsof`, `ss`, `netstat`, `tasklist`) that are used internally for port and process resolution.

## Technology Stack

- **Language**: Go 1.24+ (module declares `go 1.25.0`, CI pins `1.24.2`)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) (`github.com/spf13/cobra`)
- **Terminal Colors**: [fatih/color](https://github.com/fatih/color) (respects `NO_COLOR`)
- **YAML Parsing**: `gopkg.in/yaml.v3`
- **Build Tool**: GoReleaser (`.goreleaser.yaml`)
- **Linter**: golangci-lint (`.golangci.yml`)

## Project Structure

```
portman/
├── main.go                      # Entry point; version vars injected via ldflags
├── main_test.go                 # Sanity test for ldflag variables
├── go.mod / go.sum              # Go module files
├── cmd/                         # Cobra command definitions
│   ├── root.go                  # Root command and version setup
│   ├── list.go                  # List listening ports
│   ├── check.go                 # Check if a port is free
│   ├── kill.go                  # Kill process by port (with confirmation)
│   ├── next.go                  # Find next available port in range
│   ├── watch.go                 # Watch a port until available
│   ├── status.go                # Check project services against portman.yml
│   └── cmd_test.go              # Unit tests for commands (uses real subprocess listeners)
├── internal/                    # Internal packages
│   ├── scanner/                 # Port scanning and process lookup
│   │   ├── scanner.go           # Scanner with PortResolver interface
│   │   └── scanner_test.go      # Tests using mock resolver
│   ├── platform/                # OS-specific port/process resolution
│   │   ├── errors.go            # Shared error definitions
│   │   ├── darwin.go            # macOS: parses `lsof -i -P -n -F`
│   │   ├── darwin_test.go
│   │   ├── linux.go             # Linux: parses `ss -tulnp`
│   │   ├── linux_test.go
│   │   ├── windows.go           # Windows: parses `netstat -ano` + `tasklist /FO CSV`
│   │   └── windows_test.go
│   ├── reporter/                # Terminal output formatting
│   │   ├── reporter.go          # Tabwriter tables, colorized output
│   │   └── reporter_test.go
│   ├── registry/                # portman.yml parsing and validation
│   │   ├── registry.go
│   │   └── registry_test.go
│   └── health/                  # Health check execution
│       ├── health.go            # TCP and command checks
│       └── health_test.go
├── pkg/models/                  # Shared domain models
│   ├── port_process.go          # PortProcess struct
│   └── service_status.go        # ServiceStatus struct
├── tests/e2e/                   # End-to-end tests
│   └── e2e_test.go              # Builds binary and tests real CLI invocations
├── .github/workflows/
│   ├── ci.yml                   # CI: build, test, lint across OS matrix
│   └── release.yml              # Release: GoReleaser on version tags
├── .goreleaser.yaml             # Cross-platform release config + Homebrew tap
├── .golangci.yml                # Linter configuration
└── README.md / LICENSE / CODE_OF_CONDUCT.md
```

## Build and Test Commands

```bash
# Build the binary
go build -v ./...

# Run all unit tests
go test -v ./...

# Run with race detector (not on Windows)
go test -race ./...

# Run e2e tests (builds a temp binary named portman_e2e)
go test -v ./tests/e2e/...

# Lint (requires golangci-lint installed)
golangci-lint run --timeout=5m
```

### Release Build

Version, commit, and date are injected at link time:

```bash
go build -ldflags "-s -w -X main.version=1.0.0 -X main.commit=abc123 -X main.date=2026-04-25" -o portman .
```

Official releases are cut via GoReleaser when a `v*` tag is pushed. The release pipeline builds for:
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64 only)

## Code Organization

### Commands (`cmd/`)

Each command is a Cobra subcommand registered in `init()`:
- `list [--port <port>]` — Lists all listening ports or filters to one.
- `check <port>` — Reports free/in-use. Exits with error if in use.
- `kill <port> [--force]` — Kills the owning process. Prompts for confirmation unless `--force` is passed.
- `next [--range start-end]` — Default range is `3000-3100`.
- `watch <port> [--interval duration]` — Polls until the port is free. Default interval is `1s`.
- `status` — Looks for `portman.yml` in the current directory and prints a health table.

### Platform Abstraction (`internal/platform/`)

Platform-specific files use Go build tags (`//go:build darwin`, `//go:build linux`, `//go:build windows`).

Each platform provides a `Resolver` that implements the `PortResolver` interface:
```go
type PortResolver interface {
    GetListeningPorts() ([]models.PortProcess, error)
    GetProcessByPort(port int) (*models.PortProcess, error)
}
```

- **macOS**: Runs `lsof -i -P -n -F` and parses the machine-readable format. The `lsofRunner` variable is overridable for tests.
- **Linux**: Runs `ss -tulnp` and parses tabular output. The `ssRunner` variable is overridable for tests.
- **Windows**: Runs `netstat -ano` for ports and `tasklist /FO CSV` for process names. Both `netstatRunner` and `tasklistRunner` are overridable for tests.

`platform.ErrProcessNotFound` is the canonical error when a port has no associated process.

### Scanner (`internal/scanner/`)

`Scanner` wraps a `PortResolver` and provides higher-level operations:
- `ListPorts()` — all listening ports
- `FindProcessByPort(port)` — process for a specific port
- `IsPortFree(port)` — true if `ErrProcessNotFound`
- `FindNextAvailablePort(start, end)` — linear scan of range

`NewScanner()` uses the platform resolver; `NewScannerWithResolver(r)` allows injecting mocks.

### Reporter (`internal/reporter/`)

Uses `text/tabwriter` for aligned columns and `fatih/color` for ANSI colors.
- Respects `NO_COLOR` environment variable.
- `PrintPortTable` prints the listening-ports table.
- `PrintServiceStatusTable` prints the service-health table with green/red indicators.

### Registry (`internal/registry/`)

Parses `portman.yml` into a slice of `Service` structs. Validates that each service has a `name` and `port`.

Example `portman.yml`:
```yaml
services:
  - name: web
    port: 3000
    command: npm run dev

  - name: db
    port: 5432
    health_check: pg_isready
```

### Health (`internal/health/`)

- `TCPCheck(port)` — dials `127.0.0.1:<port>` with a 2-second timeout.
- `CommandCheck(command)` — runs `sh -c "<command>"` with a 5-second timeout, returns true on exit 0.

> **Note**: The `status` command currently uses `scanner.IsPortFree()` to determine if a service is running. `TCPCheck` is defined but not invoked by `status`.

## Testing Strategy

### Unit Tests

Every internal package has unit tests (`*_test.go` in the same package):
- **Platform tests**: Mock command runners (`lsofRunner`, `ssRunner`, `netstatRunner`, `tasklistRunner`) to test parsing logic without requiring the actual OS tools.
- **Scanner tests**: Use a `mockResolver` implementing `PortResolver` to test business logic in isolation.
- **Reporter tests**: Write to `bytes.Buffer` and assert on output strings. Test both color and no-color modes.
- **Registry tests**: Write temp YAML files, load, and validate.
- **Health tests**: Start real TCP listeners and run shell commands.

### Command Tests (`cmd/cmd_test.go`)

Tests execute commands through `executeCommand()` which captures output into buffers. Some tests start real HTTP listeners via temporary Go subprocesses (`startTestListener`) to exercise kill/check/list against actual ports.

Flag state is reset between tests to prevent leakage:
```go
killForceFlag = false
listPortFlag = 0
nextRangeFlag = "3000-3100"
watchIntervalFlag = 0
```

### E2E Tests (`tests/e2e/e2e_test.go`)

`TestMain` builds the real binary (`go build -o portman_e2e`) before running tests. Each test invokes the binary as a subprocess and asserts on stdout/stderr.

### CI Pipeline (`.github/workflows/ci.yml`)

Matrix runs on `ubuntu-latest`, `macos-latest`, `windows-latest`:
1. `go build -v ./...`
2. `go test -v ./...`
3. `go test -race ./...` (skipped on Windows)

A separate `lint` job runs `golangci-lint` on Ubuntu.

## Code Style Guidelines

- **Formatting**: `gofmt` and `goimports` (enforced by golangci-lint).
- **Linters enabled**: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`, `misspell`, `revive`.
- **Comments**: Follow standard Go conventions (exported identifiers have doc comments).
- **Error handling**: Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- **Platform separation**: Build tags (`//go:build darwin`) plus file naming (`darwin.go`, `linux.go`, `windows.go`).
- **Testability**: Platform command execution is abstracted into package-level variables (e.g., `lsofRunner`) so tests can swap them without patching.

## Security Considerations

- **Process killing**: `kill` always prompts for confirmation unless `--force` is used. Uses `os.FindProcess(pid).Kill()`.
- **Command execution**: `health.CommandCheck` runs `sh -c <command>` — the command string comes from the local `portman.yml` file, so it inherits the user's shell privileges.
- **No network calls**: The tool only dials `127.0.0.1` for TCP checks and runs local OS commands.
- **Binary releases**: Built with `CGO_ENABLED=0` for static linking.

## Notable Implementation Details

- The `PortProcess` struct includes an `Age` field (`time.Duration`), but none of the current platform resolvers populate it. It is printed in the table as a zero value.
- `lsof` on macOS sometimes exits with code 1 while still emitting usable output; the parser tolerates this when `len(out) > 0`.
- The Windows resolver gracefully degrades if `tasklist` fails: ports are still returned, but process names may be empty.
- `portman status` only looks for `portman.yml` in the current working directory.
