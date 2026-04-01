# Authentication

How agents authenticate with the desktop app over WebSocket.

## Overview

Every agent must prove its identity before streaming any data. There are no open or unauthenticated connections. The system uses a two-phase approach:

1. **Pairing** — one-time setup to register a new agent
2. **Token auth** — all subsequent connections use a signed JWT

## Pairing Flow

Pairing is how a new VPS agent is registered with your desktop app.

```
Desktop                              VPS Agent
───────                              ─────────
1. Generate pairing token
   (short-lived, one-time use)
2. Display token in UI
                                     3. User runs: agent setup --server ws://... --token <token>
                                     4. Agent saves config and connects
                                     5. Agent sends auth:pair message ──────────────►
6. Validate pairing token
7. Generate signed JWT (agent token)
8. Send auth:pair:success ◄──────────────────────
                                     9. Agent saves JWT to config
                                     10. Agent clears pairing token (one-time use)
                                     11. Session established — agent streams data
```

### Security properties

- **Pairing tokens are short-lived** — they expire after a configurable window (default: 5 minutes)
- **Pairing tokens are one-time use** — consumed on successful pairing, cannot be reused
- **Only you know the token** — it's generated on your desktop and you place it on your VPS
- **No open registration** — without a valid pairing token, the server rejects the connection

## Token Authentication Flow

After pairing, the agent stores a JWT and uses it for all future connections.

```
Desktop                              VPS Agent
───────                              ─────────
                                     1. Agent starts, reads JWT from config
                                     2. Agent connects via WebSocket
                                     3. Agent sends auth:connect ──────────────────►
4. Validate JWT (signature + expiry)
5. Send auth:connect:success ◄───────────────────
                                     6. Session established — agent streams data
```

### JWT details

| Field     | Value                                    |
|-----------|------------------------------------------|
| Algorithm | HMAC-SHA256 (HS256)                      |
| Issuer    | Desktop app instance                     |
| Subject   | Agent ID                                 |
| Expiry    | Configurable (default: 30 days)          |
| Claims    | `agentId`, `hostname`, `pairedAt`        |

The signing secret is generated once on the desktop and stored locally. It never leaves the desktop machine.

## Auto-Reconnect

When the connection drops (network issues, desktop restart, etc.), the agent automatically reconnects:

- Exponential backoff: 1s → 2s → 4s → 8s → ... capped at 60s
- Backoff resets after a successful connection
- Uses the stored JWT — no re-pairing needed
- If the JWT has expired, the server responds with `auth:connect:error` and the agent logs the issue

## Message Protocol

All messages are JSON with a `type` field for discrimination.

### Agent → Desktop

**Pairing request:**
```json
{
  "type": "auth:pair",
  "payload": {
    "pairingToken": "abc123",
    "agentId": "f7a1b2c3...",
    "hostname": "vps-prod-01"
  }
}
```

**Token authentication:**
```json
{
  "type": "auth:connect",
  "payload": {
    "agentToken": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### Desktop → Agent

**Pairing success:**
```json
{
  "type": "auth:pair:success",
  "payload": {
    "agentToken": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**Authentication success:**
```json
{
  "type": "auth:connect:success",
  "payload": {
    "agentId": "f7a1b2c3..."
  }
}
```

**Error (pairing or auth):**
```json
{
  "type": "auth:pair:error",
  "payload": {
    "message": "invalid or expired pairing token"
  }
}
```

## Auth Timeout

If a WebSocket connection does not send a valid auth message within **10 seconds**, the server closes the connection. This prevents idle or malicious connections from occupying resources.

## Agent Config File

The agent stores its state in `agent.json`:

```json
{
  "serverUrl": "ws://your-desktop-ip:9000",
  "pairingToken": "",
  "agentToken": "eyJhbGciOiJIUzI1NiIs...",
  "agentId": "f7a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5"
}
```

- `serverUrl` — desktop WebSocket server address
- `pairingToken` — set during setup, cleared after successful pairing
- `agentToken` — JWT issued by desktop, used for all subsequent connections
- `agentId` — unique identifier, auto-generated on first run

## Future Security Layers

These will be implemented as separate features:

- **TLS (wss://)** — encrypt all WebSocket traffic
- **Certificate pinning** — agents only trust your desktop's certificate
- **Rate limiting** — throttle failed auth attempts
- **IP allowlisting** — restrict which IPs can connect
- **Token rotation** — server issues fresh JWTs periodically
