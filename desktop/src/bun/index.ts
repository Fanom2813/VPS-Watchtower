import Electrobun, {
	ApplicationMenu,
	BrowserView,
	BrowserWindow,
	Updater,
	Utils,
} from "electrobun/bun";
import type { AppRPC } from "../shared/rpc-types";
import { closeDb } from "./db/database";
import * as authRepo from "./db/auth-repo";
import * as agentsRepo from "./db/agents-repo";
import * as ws from "./websocket-client";

const DEV_SERVER_PORT = 5173;
const DEV_SERVER_URL = `http://localhost:${DEV_SERVER_PORT}`;

function setupWebSocketClient(win: BrowserWindow) {
	ws.onConnect((agentId) => {
		agentsRepo.updateLastSeen(agentId);
		win.webview.rpc?.send.agentConnected({ agentId });
	});

	ws.onDisconnect((agentId) => {
		win.webview.rpc?.send.agentDisconnected({ agentId });
	});

	ws.onMessage((agentId, type, payload) => {
		win.webview.rpc?.send.agentMessage({ agentId, type, payload });
	});

	// Reconnect to all known agents on startup
	const agents = agentsRepo.getAgents();
	for (const agent of agents) {
		if (agent.token) {
			ws.connectToAgent(agent.id, agent.url, agent.token);
		}
	}
}

// Check if Vite dev server is running for HMR
async function getMainViewUrl(): Promise<string> {
	const channel = await Updater.localInfo.channel();
	if (channel === "dev") {
		try {
			await fetch(DEV_SERVER_URL, { method: "HEAD" });
			console.log(`HMR enabled: Using Vite dev server at ${DEV_SERVER_URL}`);
			return DEV_SERVER_URL;
		} catch {
			console.log(
				"Vite dev server not running. Run 'bun run dev:hmr' for HMR support.",
			);
		}
	}
	return "views://mainview/index.html";
}

// RPC handlers — bridge between UI and backend
const rpc = BrowserView.defineRPC<AppRPC>({
	handlers: {
		requests: {
			getIsSetup: () => authRepo.isSetup(),

			setupApp: () => {
				authRepo.markSetup();
			},

			readClipboard: () => {
				return Utils.clipboardReadText();
			},

			addAgent: async ({ url, pairingToken }) => {
				// Pair with the agent via WebSocket
				const result = await ws.pairWithAgent(url, pairingToken);
				const { token, agent } = result;

				// Store in database
				const row: agentsRepo.AgentRow = {
					id: agent.id,
					url,
					token,
					hostname: agent.hostname,
					label: "",
					os: agent.os,
					arch: agent.arch,
					distro: agent.distro,
					agent_version: agent.version,
					paired_at: Date.now(),
					last_seen: Date.now(),
				};
				agentsRepo.addAgent(row);

				return {
					id: row.id,
					url: row.url,
					hostname: row.hostname,
					label: row.label,
					os: row.os,
					arch: row.arch,
					distro: row.distro,
					agentVersion: row.agent_version,
					pairedAt: row.paired_at,
					lastSeen: row.last_seen,
				};
			},

			getAgents: () =>
				agentsRepo.getAgents().map((a) => ({
					id: a.id,
					url: a.url,
					hostname: a.hostname,
					label: a.label,
					os: a.os,
					arch: a.arch,
					distro: a.distro,
					agentVersion: a.agent_version,
					pairedAt: a.paired_at,
					lastSeen: a.last_seen,
				})),

			removeAgent: ({ id }) => {
				ws.disconnect(id);
				agentsRepo.removeAgent(id);
				return true;
			},

			sendToAgent: ({ agentId, message }) => {
				return ws.send(agentId, message);
			},

			broadcastToAgents: ({ message }) => {
				ws.broadcast(message);
				return true;
			},

			isAgentConnected: ({ agentId }) => {
				return ws.isConnected(agentId);
			},
		},
		messages: {},
	},
});

// Create the main application window
const url = await getMainViewUrl();

const mainWindow = new BrowserWindow({
	title: "Eyes on VPS",
	url,
	titleBarStyle: "hidden",
	styleMask: {
		Titled: false,
		Closable: true,
		Resizable: true,
		Miniaturizable: true,
		Borderless: true,
		FullSizeContentView: true,
	},
	frame: {
		width: 960,
		height: 720,
		x: 200,
		y: 200,
	},
	rpc,
});

// Connect to known agents
setupWebSocketClient(mainWindow);

// Cleanup on quit
Electrobun.events.on("before-quit", async () => {
	ws.disconnectAll();
	closeDb();
});

console.log("Eyes on VPS started!");
