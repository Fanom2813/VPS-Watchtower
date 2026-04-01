import { getDb } from "./database";

export interface AgentRow {
	id: string;
	url: string;
	token: string;
	hostname: string;
	label: string;
	os: string;
	arch: string;
	distro: string;
	agent_version: string;
	paired_at: number;
	last_seen: number;
}

export function addAgent(agent: AgentRow) {
	getDb()
		.query(
			`INSERT OR REPLACE INTO agents (id, url, token, hostname, label, os, arch, distro, agent_version, paired_at, last_seen)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		)
		.run(
			agent.id,
			agent.url,
			agent.token,
			agent.hostname,
			agent.label,
			agent.os,
			agent.arch,
			agent.distro,
			agent.agent_version,
			agent.paired_at,
			agent.last_seen,
		);
}

export function getAgents(): AgentRow[] {
	return getDb()
		.query("SELECT * FROM agents ORDER BY last_seen DESC")
		.all() as AgentRow[];
}

export function getAgent(id: string): AgentRow | null {
	const row = getDb()
		.query("SELECT * FROM agents WHERE id = ?")
		.get(id) as AgentRow | undefined;
	return row ?? null;
}

export function updateLastSeen(id: string) {
	getDb()
		.query("UPDATE agents SET last_seen = ? WHERE id = ?")
		.run(Date.now(), id);
}

export function removeAgent(id: string) {
	getDb().query("DELETE FROM agents WHERE id = ?").run(id);
}
