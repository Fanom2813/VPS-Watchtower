import type { RPCSchema } from "electrobun/bun";

export interface AgentDTO {
	id: string;
	url: string;
	hostname: string;
	label: string;
	os: string;
	arch: string;
	distro: string;
	agentVersion: string;
	pairedAt: number;
	lastSeen: number;
}

export interface UpdateInfo {
	available: boolean;
	version?: string;
	currentVersion?: string;
	error?: string;
}

export interface DownloadResult {
	success: boolean;
	ready?: boolean;
	error?: string;
}

export interface AppSettings {
	autoStart: boolean;
	startInTray: boolean;
	minimizeToTray: boolean;
	autoUpdate: boolean;
}

export type AppRPC = {
	bun: RPCSchema<{
		requests: {
			getIsSetup: { params: {}; response: boolean };
			setupApp: { params: {}; response: void };
			readClipboard: { params: {}; response: string };
			// Agent management
			addAgent: {
				params: { url: string; pairingToken: string };
				response: AgentDTO;
			};
			getAgents: { params: {}; response: AgentDTO[] };
			removeAgent: { params: { id: string }; response: boolean };
			// WebSocket operations
			sendToAgent: {
				params: { agentId: string; message: object };
				response: boolean;
			};
			broadcastToAgents: { params: { message: object }; response: boolean };
			isAgentConnected: { params: { agentId: string }; response: boolean };
			// Window controls
			closeWindow: { params: {}; response: void };
			minimizeWindow: { params: {}; response: void };
			// Updates
			checkForUpdate: { params: {}; response: UpdateInfo };
			downloadUpdate: { params: {}; response: DownloadResult };
			applyUpdate: { params: {}; response: void };
			// Settings
			getSettings: { params: {}; response: AppSettings };
			setSettings: { params: { settings: AppSettings }; response: void };
			openExternal: { params: { url: string }; response: boolean };
		};
		messages: {};
	}>;
	webview: RPCSchema<{
		requests: {};
		messages: {
			agentConnected: { agentId: string };
			agentDisconnected: { agentId: string };
			agentMessage: { agentId: string; type: string; payload: unknown };
			updateAvailable: { version: string };
			updateDownloaded: {};
		};
	}>;
};
