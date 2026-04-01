# Eyes on VPS

VPS monitoring system with a desktop dashboard and lightweight server agents.

## Architecture

```
eyes-on-vps/
├── agent/        # Go agent — single binary, runs on each VPS
└── desktop/      # Electrobun desktop app — monitoring dashboard
```

- **Agent**: Go single-binary daemon deployed to each VPS. Runs a WebSocket server that the desktop connects to. Collects auth logs, process lists, system metrics, and intrusion signals.
- **Desktop**: Electrobun (NOT Electron) app with React + Tailwind UI. Connects outbound to agents as a WebSocket client. Displays real-time alerts, processes, and system health.

**Connection direction**: Desktop → Agent (the desktop connects TO the agent on the VPS, not the other way around). This works because VPS servers have stable IPs/domains, while the desktop may be behind NAT or change networks.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Desktop runtime | Electrobun 1.16.0 (uses `electrobun/bun`, NOT Electron APIs) |
| Desktop UI | React 18 + Tailwind CSS 4 + TypeScript |
| Bundler | Vite 6 |
| Agent | Go (single static binary, cross-compiled for Linux) |
| JS runtime | Bun (desktop main process) |
| Communication | WebSocket (desktop connects to agent) |
| State management | Zustand |
| Database | SQLite (via bun:sqlite, versioned migrations) |
| JWT | jose (desktop), golang-jwt (agent) |

## References

- Electrobun docs: `/Users/omar/Documents/GitHub/llms_txt/electrobun-llms.txt`

## Key Rules

- **Electrobun is NOT Electron.** Never use Electron APIs, `electron` imports, or Electron patterns. Use `import { BrowserWindow } from "electrobun/bun"` for main process and `import { Electroview } from "electrobun/view"` for browser context.
- Use `views://` URLs to load bundled assets (e.g., `views://mainview/index.html`).
- Views must be configured in `electrobun.config.ts`.
- The main process runs on **Bun**, not Node.js.
- All styling uses Tailwind utility classes — no custom CSS files beyond the Tailwind directives.
- Use Electrobun's `Utils` API for clipboard, paths, etc. — not Node/browser APIs.
- Database migrations use `PRAGMA user_version` — never modify existing migrations, always append.

## Pairing & Auth Flow

1. **Agent setup**: `eyes-agent setup --port 9090` generates a pairing token and prints the agent URL + token.
2. **Desktop pairing**: User enters the agent URL and pairing token in the desktop app. Desktop connects via WebSocket and sends `auth:pair`.
3. **Agent validates**: Checks the pairing token, issues a JWT signed with its signing secret, responds with `auth:pair:success` + agent system info.
4. **Desktop stores**: Saves the agent record (URL, JWT, hostname, OS, etc.) in SQLite.
5. **Reconnection**: On subsequent connections, desktop sends `auth:connect` with the stored JWT. Agent verifies signature + token hash.
6. **Token regeneration**: Pairing tokens are one-time use. Run `eyes-agent token` to generate a new one.

## Desktop Development

```bash
cd desktop
bun run dev:hmr    # Recommended: Vite HMR + Electrobun dev in parallel
bun run dev        # Without HMR (manual rebuild)
bun run start      # Vite build then Electrobun dev
```

- Vite dev server runs on port **5173**.
- Main process detects Vite dev server and loads from it; falls back to bundled assets.

## Agent Development

```bash
cd agent
go run . setup --port 9090    # Configure and generate pairing token
go run . run                  # Start the WebSocket server
go run . token                # Generate a new pairing token
```

Agent binary is built via GitHub Actions — not locally. CI cross-compiles for Linux amd64/arm64.

## Project Conventions

- TypeScript strict mode everywhere.
- Agent is Go with minimal deps — prefer stdlib over third-party packages.
- Agent must be lightweight — it runs on production VPS servers with minimal overhead.
- WebSocket messages use JSON with a `type` field for message discrimination.
- Keep agent and desktop as independent packages with no shared dependencies.
- The backend (Bun process) is a thin wrapper — forward agent messages to the React frontend via RPC, handle logic in React.
