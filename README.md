# portman

> A clean CLI for managing local ports and processes. No more "Port 3000 is already in use".

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Why portman?

When you're running multiple microservices locally, managing ports becomes a chore. `portman` gives you one cross-platform command to list ports, check availability, kill processes, and even monitor ports until they're free.

## Features

- **Cross-platform** — Works identically on macOS, Linux, and Windows
- **List ports** — See all listening ports with process info in a clean table
- **Check ports** — Instantly know if a port is free or in use
- **Kill by port** — Safely terminate processes with confirmation
- **Next port** — Find the next available port in a range
- **Watch mode** — Poll a port and get notified when it's available
- **Project health** — Validate your `portman.yml` service registry

## Installation

```bash
# Homebrew (coming soon)
brew install portman

# Or download the latest release
# https://github.com/firasmosbehi/portman/releases
```

## Quick Start

```bash
# List all listening ports
portman list

# Check if port 3000 is free
portman check 3000

# Kill the process using port 8080
portman kill 8080

# Find the next available port in a range
portman next --range 3000-3010

# Watch port 3000 until it's available
portman watch 3000

# Check project services defined in portman.yml
portman status
```

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

Then run `portman status` to see if everything is healthy.

## Commands

| Command | Description |
|---------|-------------|
| `portman list` | List all listening ports with process info |
| `portman list --port 3000` | Show only port 3000 |
| `portman check <port>` | Report if port is free or in use |
| `portman kill <port>` | Find and kill process using port (with confirmation) |
| `portman next` | Suggest the next available port in a range |
| `portman watch <port>` | Monitor a port and notify when it becomes available |
| `portman status` | Check project services against `portman.yml` registry |

## Contributing

We welcome contributions! Please read our [Code of Conduct](CODE_OF_CONDUCT.md) first.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT](LICENSE)
