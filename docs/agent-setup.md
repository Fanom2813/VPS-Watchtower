# Agent Setup

How to deploy the agent on a VPS.

## Quick Start

```bash
# 1. Download the binary (from GitHub releases)
wget -O eyes-agent https://github.com/YOUR_REPO/releases/latest/download/eyes-agent-linux-amd64
chmod +x eyes-agent

# 2. Configure with your desktop's pairing token
./eyes-agent setup --server ws://YOUR_DESKTOP_IP:9000 --token PAIRING_TOKEN

# 3. Run
./eyes-agent run
```

That's it for testing. The agent connects, pairs, and starts streaming.

## Production Setup (Survives Reboots)

Add `--persist` to install as a systemd service in one step:

```bash
sudo ./eyes-agent setup --server ws://YOUR_DESKTOP_IP:9000 --token PAIRING_TOKEN --persist
```

This does everything:
- Copies the binary to `/usr/local/bin/eyes-agent`
- Creates a dedicated `eyes-agent` system user (no login, no home)
- Writes config to `/etc/eyes-on-vps/agent.json`
- Installs and enables a systemd service
- Starts the agent immediately

The agent will now:
- Start automatically on boot
- Restart automatically if it crashes (5 second delay)
- Run as a low-privilege user (not root)

## Commands

| Command | Description |
|---------|-------------|
| `eyes-agent setup --server URL --token TOKEN` | Configure the agent |
| `eyes-agent setup ... --persist` | Configure + install as systemd service |
| `eyes-agent run` | Run in foreground (for testing or manual use) |
| `eyes-agent status` | Show systemd service status |
| `sudo eyes-agent uninstall` | Stop and remove the systemd service |

## Setup Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--server` | *(required)* | Desktop app WebSocket URL |
| `--token` | *(required)* | Pairing token from the desktop app |
| `--config` | `agent.json` | Config file path |
| `--persist` | `false` | Install as a systemd service |

## Where Things Live

| What | Path |
|------|------|
| Binary | `/usr/local/bin/eyes-agent` |
| Config | `/etc/eyes-on-vps/agent.json` |
| Systemd unit | `/etc/systemd/system/eyes-agent.service` |
| Service user | `eyes-agent` (system account) |

## Managing the Service

```bash
# Check status
eyes-agent status

# Restart
sudo systemctl restart eyes-agent

# Follow logs
sudo journalctl -u eyes-agent -f

# Stop
sudo systemctl stop eyes-agent
```

## Uninstalling

```bash
sudo eyes-agent uninstall
```

This removes the binary and systemd service but **preserves the config** at `/etc/eyes-on-vps/agent.json`. Delete it manually if no longer needed:

```bash
sudo rm -rf /etc/eyes-on-vps
```

## Systemd Service Details

The service runs with security hardening:

```ini
[Service]
Type=simple
Restart=always
RestartSec=5
User=eyes-agent
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
```

- **Restart=always** — comes back after crashes or unexpected exits
- **After=network-online.target** — waits for network before starting
- **NoNewPrivileges** — cannot escalate permissions
- **ProtectSystem=strict** — filesystem is read-only except config dir
- **ProtectHome** — cannot access /home
- **PrivateTmp** — isolated /tmp

## How Pairing Works

See [authentication.md](authentication.md) for the full auth flow. In short:

1. You generate a pairing token on your desktop app
2. You pass it to `eyes-agent setup --token ...`
3. On first `run`, the agent sends the token to the desktop
4. Desktop validates it and issues a JWT
5. The agent saves the JWT — no re-pairing needed on reconnect
