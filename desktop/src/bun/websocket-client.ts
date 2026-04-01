type ConnectionHandler = (agentId: string) => void;
type MessageHandler = (agentId: string, type: string, payload: unknown) => void;

interface AgentConn {
	ws: WebSocket;
	url: string;
	token: string;
}

const connections = new Map<string, AgentConn>();
const reconnectTimers = new Map<string, Timer>();
const reconnectDelays = new Map<string, number>();

let onConnectFn: ConnectionHandler | null = null;
let onDisconnectFn: ConnectionHandler | null = null;
let onMessageFn: MessageHandler | null = null;

const BASE_DELAY = 1000;
const MAX_DELAY = 60000;

export function onConnect(handler: ConnectionHandler): void {
	onConnectFn = handler;
}

export function onDisconnect(handler: ConnectionHandler): void {
	onDisconnectFn = handler;
}

export function onMessage(handler: MessageHandler): void {
	onMessageFn = handler;
}

export interface PairResult {
	token: string;
	agent: {
		id: string;
		hostname: string;
		os: string;
		arch: string;
		distro: string;
		version: string;
	};
}

/** Connect to an agent for the first time with a pairing token. */
export function pairWithAgent(
	url: string,
	pairingToken: string,
): Promise<PairResult> {
	return new Promise((resolve, reject) => {
		let settled = false;
		const ws = new WebSocket(url);

		ws.addEventListener("open", () => {
			ws.send(
				JSON.stringify({
					type: "auth:pair",
					payload: { pairingToken },
				}),
			);
		});

		ws.addEventListener("message", (event) => {
			const msg = JSON.parse(event.data as string);

			if (msg.type === "auth:pair:success") {
				settled = true;
				const { token, agent } = msg.payload;
				connections.set(agent.id, { ws, url, token });
				installHandlers(agent.id, url, token, ws);
				onConnectFn?.(agent.id);
				resolve({ token, agent });
			} else if (msg.type === "auth:pair:error") {
				settled = true;
				ws.close();
				reject(new Error(msg.payload.message));
			}
		});

		ws.addEventListener("error", () => {
			if (!settled) {
				settled = true;
				reject(new Error("Connection failed"));
			}
		});

		ws.addEventListener("close", () => {
			if (!settled) {
				settled = true;
				reject(new Error("Connection closed"));
			}
		});
	});
}

/** Reconnect to a known agent with a stored token. */
export function connectToAgent(
	agentId: string,
	url: string,
	token: string,
): void {
	clearReconnect(agentId);

	// Close existing connection if any
	const existing = connections.get(agentId);
	if (existing) {
		existing.ws.close();
		connections.delete(agentId);
	}

	const ws = new WebSocket(url);
	let authed = false;

	ws.addEventListener("open", () => {
		ws.send(
			JSON.stringify({
				type: "auth:connect",
				payload: { token },
			}),
		);
	});

	ws.addEventListener("message", (event) => {
		const msg = JSON.parse(event.data as string);

		if (!authed) {
			if (msg.type === "auth:connect:success") {
				authed = true;
				reconnectDelays.delete(agentId);
				connections.set(agentId, { ws, url, token });
				installHandlers(agentId, url, token, ws);
				onConnectFn?.(agentId);
			} else if (msg.type === "auth:connect:error") {
				console.log(`[WS] Auth failed for ${agentId}`);
				ws.close();
				// Don't auto-reconnect on auth errors
			}
			return;
		}

		if (msg.type) {
			onMessageFn?.(agentId, msg.type, msg.payload);
		}
	});

	ws.addEventListener("error", () => {});

	ws.addEventListener("close", () => {
		if (authed) {
			connections.delete(agentId);
			onDisconnectFn?.(agentId);
		}
		scheduleReconnect(agentId, url, token);
	});
}

function installHandlers(
	agentId: string,
	url: string,
	token: string,
	ws: WebSocket,
): void {
	ws.onmessage = (event) => {
		const msg = JSON.parse(event.data as string);
		if (msg.type) {
			onMessageFn?.(agentId, msg.type, msg.payload);
		}
	};

	ws.onclose = () => {
		connections.delete(agentId);
		onDisconnectFn?.(agentId);
		scheduleReconnect(agentId, url, token);
	};

	ws.onerror = () => {};
}

function scheduleReconnect(
	agentId: string,
	url: string,
	token: string,
): void {
	const delay = reconnectDelays.get(agentId) ?? BASE_DELAY;

	const timer = setTimeout(() => {
		reconnectTimers.delete(agentId);
		connectToAgent(agentId, url, token);
	}, delay);

	reconnectTimers.set(agentId, timer);
	reconnectDelays.set(agentId, Math.min(delay * 2, MAX_DELAY));
}

function clearReconnect(agentId: string): void {
	const timer = reconnectTimers.get(agentId);
	if (timer) {
		clearTimeout(timer);
		reconnectTimers.delete(agentId);
	}
}

export function send(agentId: string, message: object): boolean {
	const conn = connections.get(agentId);
	if (!conn || conn.ws.readyState !== WebSocket.OPEN) return false;
	conn.ws.send(JSON.stringify(message));
	return true;
}

export function broadcast(message: object): void {
	const data = JSON.stringify(message);
	for (const conn of connections.values()) {
		if (conn.ws.readyState === WebSocket.OPEN) {
			conn.ws.send(data);
		}
	}
}

export function isConnected(agentId: string): boolean {
	const conn = connections.get(agentId);
	return conn?.ws.readyState === WebSocket.OPEN;
}

export function disconnect(agentId: string): void {
	clearReconnect(agentId);
	reconnectDelays.delete(agentId);
	const conn = connections.get(agentId);
	if (conn) {
		conn.ws.close();
		connections.delete(agentId);
	}
}

export function disconnectAll(): void {
	for (const id of [...connections.keys()]) {
		disconnect(id);
	}
}
