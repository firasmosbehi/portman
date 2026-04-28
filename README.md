# portman

> A clean CLI for managing local ports and processes. No more "Port 3000 is already in use".

[![CI](https://github.com/firasmosbehi/portman/actions/workflows/ci.yml/badge.svg)](https://github.com/firasmosbehi/portman/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Why portman?

When you're running multiple microservices locally, managing ports becomes a chore. `portman` gives you one cross-platform command to list ports, check availability, kill processes, and even monitor ports until they're free.

## Features

- **Cross-platform** — Works identically on macOS, Linux, and Windows
- **List ports** — See all listening ports with process info (including process age) in a clean, colorized table. Export as JSON too.
- **Find ports** — Search listening ports by process name or PID
- **Check ports** — Instantly know if a port is free or in use
- **Kill by port** — Safely terminate processes with confirmation (`--force` to skip)
- **Next port** — Find the next available port in a range
- **Watch mode** — Poll a port and get live updates until it's available
- **Project health** — Validate your `portman.yml` service registry
- **Scaffold configs** — Generate a `portman.yml` with `portman init`

## Installation

### Homebrew

```bash
brew tap firasmosbehi/tap
brew install portman
```

### Pre-built binaries

Download the latest release for your platform from the [releases page](https://github.com/firasmosbehi/portman/releases).

### From source

```bash
go install github.com/firasmosbehi/portman@latest
```

## Quick Start

```bash
# List all listening ports
portman list

# Show only a specific port
portman list --port 3000

# Export as JSON
portman list --format json

# Find all ports used by a process
portman find node

# Find ports by exact PID
portman find --pid 4521

# Check if port 3000 is free
portman check 3000

# Kill the process using port 8080 (with confirmation)
portman kill 8080

# Kill without confirmation
portman kill 8080 --force

# Find the next available port in a range
portman next --range 3000-3010

# Watch port 3000 until it's available (polls every 1s)
portman watch 3000

# Watch with a custom polling interval
portman watch 3000 --interval 5s

# Check project services defined in portman.yml
portman status

# Generate a sample portman.yml
portman init

# Generate an empty portman.yml
portman init --blank
```

## Commands

| Command | Description |
|---------|-------------|
| `portman list` | List all listening ports with process info |
| `portman list --port <port>` | Show only the specified port |
| `portman list --format json` | Output as JSON instead of table |
| `portman find <process>` | Find ports by process name (substring match) |
| `portman find --pid <pid>` | Find ports by exact PID |
| `portman check <port>` | Report if a port is free or in use |
| `portman kill <port>` | Find and kill the process using a port |
| `portman kill <port> --force` | Kill without confirmation |
| `portman next` | Suggest the next available port (default range: 3000-3100) |
| `portman next --range <start-end>` | Scan a custom range |
| `portman watch <port>` | Monitor a port until it becomes available |
| `portman watch <port> --interval <duration>` | Set custom polling interval |
| `portman status` | Check project services against `portman.yml` |
| `portman init` | Generate a sample `portman.yml` |
| `portman init --blank` | Generate an empty `portman.yml` |
| `portman init --force` | Overwrite existing `portman.yml` |

## Project Registry

Create a `portman.yml` in your project root to declare expected services:

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

Then run `portman status` to see if everything is healthy:

```
Project Services

SERVICE  EXPECTED  ACTUAL  STATUS
───────────────────────────────────
web      3000      3000    ✓ running
api      3001      3001    ✓ running
db       5432      5432    ✓ healthy
cache    6379      6379    ✓ running

All services healthy.
```

## How it works

`portman` uses platform-native tools under the hood:

- **macOS**: `lsof -i -P -n -F` (machine-readable format)
- **Linux**: `ss -tulnp`
- **Windows**: `netstat -ano` + `tasklist /FO CSV`

All parsing is done internally — you get the same clean output on every OS.

## Contributing

We welcome contributions! Please read our [Code of Conduct](CODE_OF_CONDUCT.md) first.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT](LICENSE)
