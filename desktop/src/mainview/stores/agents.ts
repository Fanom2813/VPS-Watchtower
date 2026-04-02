import { create } from "zustand";
import { rpc } from "@/lib/rpc";
import type { AgentDTO } from "../../shared/rpc-types";

export type AgentStatus = "online" | "offline";

// Collector data types
export interface SystemMetrics {
	cpuPercent: number;
	memTotal: number;
	memUsed: number;
	memPercent: number;
	diskTotal: number;
	diskUsed: number;
	diskPercent: number;
	uptime: number;
	loadAvg: [number, number, number];
	netRxBytes: number;
	netTxBytes: number;
	timestamp: number;
}

export interface ProcessInfo {
	pid: number;
	ppid: number;
	name: string;
	binPath: string;
	cmdline: string;
	state: string;
	user: string;
	class: "kernel" | "system" | "unknown";
	cpuTime: number;
	memRss: number;
}

export interface ProcessList {
	processes: ProcessInfo[];
	total: number;
	unknown: number;
	timestamp: number;
}

export interface AuthEntry {
	timestamp: string;
	type: "login_success" | "login_failed" | "disconnect" | "invalid_user";
	user: string;
	source: string;
	method: string;
	raw: string;
}

export interface AuthLogPayload {
	entries: AuthEntry[];
	timestamp: number;
}

export interface NetConnection {
	localAddr: string;
	localPort: number;
	remoteAddr: string;
	remotePort: number;
	state: string;
	pid?: number;
}

export interface NetworkPayload {
	connections: NetConnection[];
	listening: NetConnection[];
	total: number;
	timestamp: number;
}

export interface DockerContainer {
	id: string;
	name: string;
	image: string;
	state: string;
	status: string;
	created: number;
}

export interface DockerPayload {
	available: boolean;
	containers: DockerContainer[];
	total: number;
	timestamp: number;
}

export interface IntrusionAlert {
	type: "brute_force" | "priv_escalation" | "port_scan";
	severity: "low" | "medium" | "high" | "critical";
	source: string;
	detail: string;
	count: number;
	timestamp: number;
}

export interface IntrusionPayload {
	alerts: IntrusionAlert[];
	timestamp: number;
}

export interface OutboundConnection {
	remoteAddr: string;
	remotePort: number;
	localPort: number;
	firstSeen: number;
	new: boolean;
}

export interface OutboundPayload {
	connections: OutboundConnection[];
	newDestinations: number;
	total: number;
	timestamp: number;
}

export interface TamperPayload {
	binaryPath: string;
	hash: string;
	modified: boolean;
	timestamp: number;
}

export interface CronEntry {
	source: string;
	schedule: string;
	command: string;
	user: string;
}

export interface CronPayload {
	jobs: CronEntry[];
	total: number;
	timestamp: number;
}

export interface ServiceInfo {
	name: string;
	path: string;
	type: "service" | "timer" | "socket";
	enabled: boolean;
	modified: number;
	snippet: string;
}

export interface ServicesPayload {
	services: ServiceInfo[];
	total: number;
	timestamp: number;
}

export interface FileState {
	path: string;
	hash: string;
	modTime: number;
	size: number;
	changed: boolean;
	missing: boolean;
}

export interface FileIntegrityPayload {
	files: FileState[];
	changes: number;
	timestamp: number;
}

export interface SensitiveAccess {
	pid: number;
	process: string;
	file: string;
	reason: string;
	user: string;
}

export interface SensitiveAccessPayload {
	accesses: SensitiveAccess[];
	total: number;
	timestamp: number;
}

// Agent collector data
export interface AgentCollectorData {
	system?: SystemMetrics;
	processes?: ProcessList;
	authlog?: AuthLogPayload;
	network?: NetworkPayload;
	docker?: DockerPayload;
	intrusion?: IntrusionPayload;
	outbound?: OutboundPayload;
	tamper?: TamperPayload;
	cron?: CronPayload;
	services?: ServicesPayload;
	fileIntegrity?: FileIntegrityPayload;
	sensitiveAccess?: SensitiveAccessPayload;
}

export interface Agent extends AgentDTO {
	status: AgentStatus;
	collectorData: AgentCollectorData;
}

interface AgentsState {
	agents: Agent[];
	loading: boolean;

	loadAgents: () => Promise<void>;
	addAgent: (url: string, pairingToken: string) => Promise<Agent>;
	removeAgent: (id: string) => Promise<void>;
	setOnline: (id: string) => void;
	setOffline: (id: string) => void;
	updateCollectorData: (id: string, type: string, payload: unknown) => void;
}

export const useAgentsStore = create<AgentsState>()((set, get) => ({
	agents: [],
	loading: true,

	loadAgents: async () => {
		set({ loading: true });
		const rows = await rpc.request.getAgents({});

		// Check actual connection status for each agent
		const agentsWithStatus = await Promise.all(
			rows.map(async (a) => {
				const isConnected = await rpc.request.isAgentConnected({ agentId: a.id });
				return {
					...a,
					status: isConnected ? ("online" as const) : ("offline" as const),
					collectorData: {},
				};
			}),
		);

		set({
			agents: agentsWithStatus,
			loading: false,
		});
	},

	addAgent: async (url, pairingToken) => {
		const dto = await rpc.request.addAgent({ url, pairingToken });
		const agent: Agent = { ...dto, status: "online", collectorData: {} };
		set({ agents: [agent, ...get().agents] });
		return agent;
	},

	removeAgent: async (id) => {
		await rpc.request.removeAgent({ id });
		set({ agents: get().agents.filter((a) => a.id !== id) });
	},

	setOnline: (id) =>
		set({
			agents: get().agents.map((a) =>
				a.id === id ? { ...a, status: "online" } : a,
			),
		}),

	setOffline: (id) =>
		set({
			agents: get().agents.map((a) =>
				a.id === id ? { ...a, status: "offline" } : a,
			),
		}),

	updateCollectorData: (id, type, payload) =>
		set({
			agents: get().agents.map((a) => {
				if (a.id !== id) return a;
				const data = { ...a.collectorData };
				switch (type) {
					case "metrics:system":
						data.system = payload as SystemMetrics;
						break;
					case "metrics:processes":
						data.processes = payload as ProcessList;
						break;
					case "metrics:authlog":
						data.authlog = payload as AuthLogPayload;
						break;
					case "metrics:network":
						data.network = payload as NetworkPayload;
						break;
					case "metrics:docker":
						data.docker = payload as DockerPayload;
						break;
					case "metrics:intrusion":
						data.intrusion = payload as IntrusionPayload;
						break;
					case "metrics:outbound":
						data.outbound = payload as OutboundPayload;
						break;
					case "metrics:tamper":
						data.tamper = payload as TamperPayload;
						break;
					case "metrics:cron":
						data.cron = payload as CronPayload;
						break;
					case "metrics:services":
						data.services = payload as ServicesPayload;
						break;
					case "metrics:file_integrity":
						data.fileIntegrity = payload as FileIntegrityPayload;
						break;
					case "metrics:sensitive_access":
						data.sensitiveAccess = payload as SensitiveAccessPayload;
						break;
				}
				return { ...a, collectorData: data };
			}),
		}),
}));
