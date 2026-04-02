# VPS Watchtower

> ⚠️ **Under Active Development** — This project is a work in progress. Core features are being implemented and the API may change. Not yet ready for production use.

**Lightweight VPS monitoring and security surveillance with a real-time desktop dashboard.**

![Version](https://img.shields.io/badge/version-1.0.0--dev-orange)
![License](https://img.shields.io/badge/license-MIT-green)

VPS Watchtower helps you monitor multiple VPS instances from a beautiful desktop application. Deploy lightweight agents to your servers and get real-time insights into system health, security events, processes, and more.

![Dashboard Preview](./docs/screenshot.png)

## Features

### 🖥️ Desktop Dashboard
- **Real-time monitoring** - Live updates from all connected agents
- **Multi-server support** - Manage unlimited VPS instances
- **Beautiful UI** - Modern, dark-themed interface built with React + Tailwind
- **Auto-updates** - Built-in updater keeps you on the latest version

### 🛡️ Security Monitoring
- **Auth log analysis** - Detect failed logins and suspicious access
- **Intrusion detection** - Monitor for unauthorized access attempts
- **Sensitive file monitoring** - Track changes to critical system files
- **Tamper detection** - Alert on system modifications

### 📊 System Insights
- **Process monitoring** - See running processes and resource usage
- **Network connections** - Track inbound/outbound connections
- **Docker containers** - Monitor container status and events
- **Cron jobs** - View scheduled tasks
- **Systemd services** - Check service health
- **System metrics** - CPU, memory, disk, and network stats

## Architecture

```
┌─────────────────┐      WebSocket      ┌─────────────────┐
│   Desktop App   │ ◄─────────────────► │   Agent (VPS)   │
│  (Electrobun)   │   Secure Connection │   (Go binary)   │
└─────────────────┘                     └─────────────────┘
     Your PC                                  Linux
   macOS/Windows                              Any VPS
```

- **Agent**: Single Go binary (~10MB) deployed to each VPS. Runs as a systemd service, collects metrics, and exposes a WebSocket server.
- **Desktop**: Cross-platform desktop app (macOS, Windows, Linux) built with Electrobun. Connects to agents and displays real-time data.

**Connection flow**: Desktop → Agent (outbound from desktop to VPS). This works because VPS servers have stable IPs while desktops may be behind NAT.

## Quick Start

### 1. Install the Desktop App

```bash
cd desktop
bun install
bun run dev:hmr    # Development
bun run build      # Production build
```

### 2. Deploy an Agent to Your VPS

```bash
# On your VPS (Linux)
cd agent
go build -o eyes-agent .
./eyes-agent setup --port 9090
```

This outputs:
```
Agent URL:      ws://YOUR_VPS_IP:9090/ws
Pairing Token:  eyJhbGc...
```

### 3. Connect Desktop to Agent

1. Open the desktop app
2. Click "Add Server"
3. Enter the Agent URL and Pairing Token
4. Click "Connect"

### 4. Run the Agent (Production)

```bash
# Install as systemd service (runs on boot)
sudo ./eyes-agent setup --port 9090 --persist

# Useful commands
sudo systemctl status eyes-agent
sudo journalctl -u eyes-agent -f
sudo eyes-agent uninstall
```

## Project Structure

```
eyes-on-vps/
├── agent/              # Go agent - runs on each VPS
│   ├── cmd.go          # CLI entry point
│   └── internal/
│       ├── auth/       # JWT authentication
│       ├── collector/  # System data collectors
│       ├── config/     # Agent configuration
│       ├── protocol/   # WebSocket message protocol
│       ├── service/    # Systemd service installation
│       ├── sysinfo/    # System information utilities
│       └── transport/  # WebSocket server
│
├── desktop/            # Electrobun desktop app
│   ├── src/
│   │   ├── bun/        # Main process (Bun runtime)
│   │   ├── mainview/   # React UI components
│   │   └── shared/     # Shared types and RPC schema
│   └── build/          # Platform-specific builds
│
└── docs/               # Documentation and assets
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Desktop Runtime | [Electrobun](https://electrobun.dev) (Bun-based, NOT Electron) |
| Desktop UI | React 18 + Tailwind CSS 4 + TypeScript |
| State Management | Zustand |
| Database | SQLite (bun:sqlite) |
| Agent | Go 1.21+ (single static binary) |
| Communication | WebSocket with JWT authentication |
| Build | Vite 6 (frontend), Go build (agent) |

## Development

### Desktop Development

```bash
cd desktop

# Install dependencies
bun install

# Development with hot reload (recommended)
bun run dev:hmr

# Build for production
bun run build

# Build release for your platform
bun run build:prod
```

### Agent Development

```bash
cd agent

# Run in development mode
go run . setup --port 9090
go run . run

# Build for production
GOOS=linux GOARCH=amd64 go build -o eyes-agent .
```

### Cross-compile Agent for All Platforms

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o eyes-agent-linux-amd64 .

# Linux arm64
GOOS=linux GOARCH=arm64 go build -o eyes-agent-linux-arm64 .
```

## Configuration

### Agent Configuration (agent.json)

```json
{
  "agentId": "unique-id",
  "port": 9090,
  "pairingToken": "generated-token"
}
```

### Desktop Settings

- **Auto-start**: Launch on system login
- **Start in tray**: Minimize to system tray on launch
- **Minimize to tray**: Keep running in tray when minimized
- **Auto-update**: Automatically download and install updates

## Security

- **JWT Authentication**: Agents issue JWTs signed with a secret known only to the agent
- **One-time pairing tokens**: Tokens are invalidated after first use
- **Secure WebSocket**: All communication is encrypted (use wss:// in production)
- **No inbound connections**: Desktop initiates all connections (works behind NAT)

### Best Practices

1. **Use HTTPS/WSS**: Deploy with a reverse proxy (nginx, Caddy) for TLS
2. **Firewall rules**: Only allow connections from trusted IPs
3. **Rotate tokens**: Run `eyes-agent token` to generate new pairing tokens
4. **Restrict port**: Use firewall to limit access to the agent port

## System Requirements

### Agent (VPS)
- Linux (amd64 or arm64)
- systemd (for service management)
- ~10MB disk space
- Minimal CPU/memory overhead

### Desktop
- macOS, Windows, or Linux
- ~200MB disk space
- Modern GPU for smooth UI

## Troubleshooting

### Agent won't connect
- Check firewall rules: `sudo ufw allow 9090`
- Verify the agent is running: `sudo systemctl status eyes-agent`
- Check logs: `sudo journalctl -u eyes-agent -f`

### Desktop won't start
- Delete the database: `rm ~/Library/Application\ Support/eyes-on-vps/app.db` (macOS)
- Rebuild the app: `bun run build`

### Pairing fails
- Ensure the token hasn't been used (run `eyes-agent token` to generate a new one)
- Check that the agent URL is correct (use the IP shown in setup)

## Roadmap

- [ ] Mobile app (iOS/Android)
- [ ] Push notifications for critical alerts
- [ ] Historical data and charts
- [ ] Plugin system for custom collectors
- [ ] Team collaboration features
- [ ] Cloud sync for desktop settings

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Electrobun](https://electrobun.dev) - Fast desktop app runtime
- UI components inspired by [shadcn/ui](https://ui.shadcn.com)
- Icons by [Lucide](https://lucide.dev)

---

**Made with ❤️ for VPS administrators and security enthusiasts**
