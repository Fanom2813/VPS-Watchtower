import { create } from "zustand";
import { rpc } from "@/lib/rpc";

interface AuthState {
	isSetup: boolean;
	loading: boolean;

	checkSetup: () => Promise<void>;
	setupApp: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()((set) => ({
	isSetup: false,
	loading: true,

	checkSetup: async () => {
		set({ loading: true });
		const isSetup = await rpc.request.getIsSetup({});
		set({ isSetup, loading: false });
	},

	setupApp: async () => {
		await rpc.request.setupApp({});
		set({ isSetup: true });
	},
}));
