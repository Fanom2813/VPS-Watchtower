import { Electroview } from "electrobun/view";
import type { AppRPC } from "../../shared/rpc-types";

const rpcDef = Electroview.defineRPC<AppRPC>({
	handlers: {
		requests: {},
		messages: {
			agentConnected: (payload) => {
				window.dispatchEvent(
					new CustomEvent("agent:connected", { detail: payload.agentId }),
				);
			},
			agentDisconnected: (payload) => {
				window.dispatchEvent(
					new CustomEvent("agent:disconnected", { detail: payload.agentId }),
				);
			},
			agentMessage: (payload) => {
				window.dispatchEvent(
					new CustomEvent("agent:message", {
						detail: {
							agentId: payload.agentId,
							type: payload.type,
							payload: payload.payload,
						},
					}),
				);
			},
		},
	},
});

export const electroview = new Electroview({ rpc: rpcDef });

// Non-null accessor — RPC is always defined when Electroview is initialized.
export const rpc = electroview.rpc!;
