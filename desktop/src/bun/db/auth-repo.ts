import { getDb } from "./database";

export function getConfig(key: string): string | null {
	const row = getDb()
		.query("SELECT value FROM config WHERE key = ?")
		.get(key) as { value: string } | undefined;
	return row?.value ?? null;
}

export function setConfig(key: string, value: string) {
	getDb()
		.query("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)")
		.run(key, value);
}

export function isSetup(): boolean {
	return getConfig("is_setup") === "true";
}

export function markSetup() {
	setConfig("is_setup", "true");
}
