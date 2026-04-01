import { BrowserRouter, Routes, Route, Navigate } from "react-router";
import { AppLayout } from "@/components/layout/app-layout";
import { AuthPage } from "@/pages/auth";
import { OverviewPage } from "@/pages/overview";
import { AgentDetailsPage } from "@/pages/agent-details";
import { useAgentsStore } from "@/stores/agents";
import { useAuthStore } from "@/stores/auth";
import { useEffect } from "react";

function AppRoot() {
	const { agents, loading: agentsLoading, loadAgents } = useAgentsStore();
	const {
		isSetup,
		loading: authLoading,
		checkSetup,
		setupApp,
	} = useAuthStore();

	useEffect(() => {
		checkSetup();
		loadAgents();
	}, []);

	// Ensure app is set up (generates signing secret on first run)
	useEffect(() => {
		if (!authLoading && !isSetup) {
			setupApp();
		}
	}, [authLoading, isSetup]);

	if (authLoading || agentsLoading) return null;

	if (agents.length === 0) {
		return <Navigate to="/add" replace />;
	}

	return <Navigate to="/overview" replace />;
}

function App() {
	return (
		<BrowserRouter>
			<AppLayout>
				<Routes>
					<Route path="/" element={<AppRoot />} />
					<Route path="/add" element={<AuthPage />} />
					<Route path="/overview" element={<OverviewPage />} />
					<Route path="/agents/:id" element={<AgentDetailsPage />} />
					<Route path="*" element={<Navigate to="/" replace />} />
				</Routes>
			</AppLayout>
		</BrowserRouter>
	);
}

export default App;
