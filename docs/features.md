# Features

## Shareable Access Links

- Generate invite links that grant access to a specific server's dashboard
- Recipients click the link and get logged in — no credential sharing needed
- Useful for teams or friends collaborating on the same VPS
- Links can be revoked or set to expire

## Agent Security

- Agents authenticate with the desktop app using a password or token-based system
- No open/unauthenticated connections — every agent must prove identity before streaming data
- Secure pairing process when adding a new VPS

## Agent Notifications (Remote)

- Agents can send alerts through external channels when critical events occur
- Supported platforms:
  - Telegram
  - WhatsApp
  - Email
- Configurable per-agent — choose which channels to enable and what triggers them
- Works independently of the desktop app (alerts still go out even if the app is closed)

## Desktop App — System Tray

- App minimizes to the system tray
- Tray icon click shows a quick stats popup (CPU, memory, disk usage)
- Runs quietly in the background without taking up taskbar space

## Desktop App — Local Notifications

- Desktop push notifications for important events (intrusion attempts, agent disconnects, high resource usage)
- Notifications appear even when the app is minimized to tray
- Click a notification to jump straight to the relevant server/alert

## Live Monitoring

- Real-time system metrics — CPU, memory, disk, network per server
- Process explorer — see what's running on each VPS
- Auth log viewer — SSH attempts, logins, failed passwords
- Historical graphs — track metrics over time, not just live numbers

## Multi-Server Management

- Server overview — all VPS instances at a glance with online/offline status
- Add and remove servers easily from the dashboard
- Per-server detail view with dedicated metrics, processes, and logs

## Intrusion Detection

- Brute-force SSH attempt detection
- Unknown/suspicious login alerts
- Privilege escalation detection
- Port scan detection

## Agent Reliability

- Auto-reconnect when connection to the desktop app drops
- Self-update mechanism for pushing new agent versions
- Configurable monitoring intervals (how often the agent reports data)

## Connection Security

- TLS/encryption on WebSocket connections between agent and desktop
- Certificate pinning so agents only communicate with your app, not a MITM

## Access Control

- Role-based permissions on shared links (view-only vs full access)
- Session timeouts — shared links auto-expire after inactivity
- Audit log — who accessed what server and when

## Agent Hardening

- Agent runs as a dedicated low-privilege user, not root
- Rate limiting on auth attempts to the desktop app
- IP allowlisting — only accept agent connections from known IPs

## Data Protection

- Sensitive data (tokens, passwords, notification API keys) encrypted at rest
- Logs redacted for passwords and secrets before display

## Tamper Detection

- Agent integrity check — detect if the agent binary was modified on the VPS
- Alert if the agent process is killed or stops unexpectedly

## Process Whitelisting

- Whitelist known/trusted processes and services per server (mark them green)
- Any new unknown process triggers an immediate alert
- Detects crypto miners, backdoors, or any unauthorized software
- System processes auto-recognized — alerts only for non-standard processes
- Dashboard shows a clear distinction between whitelisted, system, and unknown processes

## Outbound Traffic Monitoring

- Track all outbound network connections — destination IP, port, protocol, and volume
- Alert on unexpected outbound traffic (data exfiltration, reverse shells, C2 callbacks)
- See where data is going, how much, and when
- Flag new outbound destinations that haven't been seen before
- Useful for detecting stolen credentials being sent out or unauthorized data transfers

## Environment Variable Protection

- Monitor access and changes to environment variables on the VPS
- Alert if any env var is read by an unauthorized process
- Alert if env vars are modified, added, or deleted
- Track which process touched which env var and when
- Protects API keys, database credentials, and secrets stored in the environment
