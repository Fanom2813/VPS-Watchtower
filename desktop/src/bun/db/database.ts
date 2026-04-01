import { Database } from "bun:sqlite";
import { Utils } from "electrobun/bun";
import { mkdirSync } from "fs";
import { join } from "path";

let db: Database;

export function getDb(): Database {
	if (!db) {
		const dir = Utils.paths.userData;
		mkdirSync(dir, { recursive: true });

		const dbPath = join(dir, "eyes-on-vps.db");
		db = new Database(dbPath);
		db.run("PRAGMA journal_mode = WAL");
		db.run("PRAGMA foreign_keys = ON");

		migrate(db);
		console.log(`Database opened at ${dbPath}`);
	}
	return db;
}

export function closeDb() {
	if (db) {
		db.close();
	}
}

function getVersion(db: Database): number {
	const row = db.query("PRAGMA user_version").get() as {
		user_version: number;
	};
	return row.user_version;
}

function setVersion(db: Database, version: number) {
	db.run(`PRAGMA user_version = ${version}`);
}

/**
 * Each migration function brings the database from version N-1 to N.
 * Migrations are run in order and wrapped in a transaction.
 * Never modify an existing migration — always add a new one.
 */
const migrations: ((db: Database) => void)[] = [
	// v1: initial schema (config + agents with old columns)
	(db) => {
		db.run(`
			CREATE TABLE IF NOT EXISTS config (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL
			)
		`);

		db.run(`
			CREATE TABLE IF NOT EXISTS agents (
				id TEXT PRIMARY KEY,
				hostname TEXT NOT NULL DEFAULT '',
				label TEXT NOT NULL DEFAULT '',
				os TEXT NOT NULL DEFAULT '',
				arch TEXT NOT NULL DEFAULT '',
				distro TEXT NOT NULL DEFAULT '',
				agent_version TEXT NOT NULL DEFAULT '',
				paired_at INTEGER NOT NULL DEFAULT 0,
				last_seen INTEGER NOT NULL DEFAULT 0,
				token_hash TEXT NOT NULL DEFAULT ''
			)
		`);

		db.run(`
			CREATE TABLE IF NOT EXISTS pairing_tokens (
				token TEXT PRIMARY KEY,
				label TEXT NOT NULL DEFAULT '',
				created_at INTEGER NOT NULL,
				expires_at INTEGER NOT NULL
			)
		`);
	},

	// v2: flip architecture — desktop connects to agent
	//     agents get url + token columns, drop pairing_tokens
	(db) => {
		db.run("DROP TABLE IF EXISTS pairing_tokens");

		// Recreate agents with new schema
		db.run("ALTER TABLE agents RENAME TO agents_old");
		db.run(`
			CREATE TABLE agents (
				id TEXT PRIMARY KEY,
				url TEXT NOT NULL,
				token TEXT NOT NULL DEFAULT '',
				hostname TEXT NOT NULL DEFAULT '',
				label TEXT NOT NULL DEFAULT '',
				os TEXT NOT NULL DEFAULT '',
				arch TEXT NOT NULL DEFAULT '',
				distro TEXT NOT NULL DEFAULT '',
				agent_version TEXT NOT NULL DEFAULT '',
				paired_at INTEGER NOT NULL DEFAULT 0,
				last_seen INTEGER NOT NULL DEFAULT 0
			)
		`);
		db.run("DROP TABLE agents_old");
	},
];

function migrate(db: Database) {
	const current = getVersion(db);
	const target = migrations.length;

	if (current >= target) return;

	db.run("BEGIN");
	try {
		for (let i = current; i < target; i++) {
			migrations[i](db);
		}
		setVersion(db, target);
		db.run("COMMIT");
		console.log(`Database migrated from v${current} to v${target}`);
	} catch (err) {
		db.run("ROLLBACK");
		throw err;
	}
}
