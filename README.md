# portwatch

Lightweight CLI daemon that monitors open ports and alerts on unexpected changes.

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

## Usage

Start the daemon with default settings (scans every 60 seconds):

```bash
portwatch start
```

Specify a custom scan interval and alert via webhook:

```bash
portwatch start --interval 30 --webhook https://hooks.example.com/alert
```

Define a baseline of expected open ports to suppress known services:

```bash
portwatch start --allow 22,80,443
```

When an unexpected port opens or closes, `portwatch` prints an alert to stdout and optionally fires a webhook:

```
[ALERT] 2024-11-03 14:22:01 — new port detected: 8080 (PID 3821, process: node)
[ALERT] 2024-11-03 14:25:14 — port closed: 443
```

### Commands

| Command | Description |
|---|---|
| `start` | Start the monitoring daemon |
| `snapshot` | Print current open ports and exit |
| `version` | Print version information |

## Configuration

`portwatch` can be configured via a YAML file at `~/.portwatch.yaml`:

```yaml
interval: 60
allow:
  - 22
  - 80
  - 443
webhook: https://hooks.example.com/alert
```

## License

MIT — see [LICENSE](LICENSE) for details.