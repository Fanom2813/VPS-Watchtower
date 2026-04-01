import { create } from "zustand";
import { rpc } from "@/lib/rpc";
import type { AgentDTO } from "../../shared/rpc-types";

export type AgentStatus = "online" | "offline";

export interface Agent extends AgentDTO {
	status: AgentStatus;
}

interface AgentsState {
	agents: Agent[];
	loading: boolean;

	loadAgents: () => Promise<void>;
	addAgent: (url: string, pairingToken: string) => Promise<Agent>;
	removeAgent: (id: string) => Promise<void>;
	setOnline: (id: string) => void;
	setOffline: (id: string) => void;
}

export const useAgentsStore = create<AgentsState>()((set, get) => ({
	agents: [],
	loading: true,

	loadAgents: async () => {
		set({ loading: true });
		const rows = await rpc.request.getAgents({});
		set({
			agents: rows.map((a) => ({ ...a, status: "offline" as const })),
			loading: false,
		});
	},

	addAgent: async (url, pairingToken) => {
		const dto = await rpc.request.addAgent({ url, pairingToken });
		const agent: Agent = { ...dto, status: "online" };
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
}));
