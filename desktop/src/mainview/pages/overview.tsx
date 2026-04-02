import { useEffect } from "react";
import { useNavigate } from "react-router";
import { Monitor, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AgentListItem } from "@/components/agents/agent-list-item";
import { useAgentsStore } from "@/stores/agents";
import { PageLayout } from "@/components/layout/page-layout";

export function OverviewPage() {
	const navigate = useNavigate();
	const { agents, loadAgents, removeAgent, setOnline, setOffline } =
		useAgentsStore();

	useEffect(() => {
		loadAgents();

		const onConnect = (e: Event) => {
			const agentId = (e as CustomEvent).detail;
			setOnline(agentId);
		};
		const onDisconnect = (e: Event) => {
			const agentId = (e as CustomEvent).detail;
			setOffline(agentId);
		};

		window.addEventListener("agent:connected", onConnect);
		window.addEventListener("agent:disconnected", onDisconnect);
		return () => {
			window.removeEventListener("agent:connected", onConnect);
			window.removeEventListener("agent:disconnected", onDisconnect);
		};
	}, []);

	const handleRemove = async (id: string) => {
		await removeAgent(id);
		if (agents.length <= 1) {
			navigate("/add");
		}
	};

	const actions = (
		<Button
			variant="outline"
			className="uppercase tracking-wider text-label"
			onClick={() => navigate("/add")}
		>
			<Plus className="w-3.5 h-3.5 mr-2" />
			Add Server
		</Button>
	);

	return (
		<PageLayout
			title="Servers"
			subtitle="Your connected VPS instances"
			actions={actions}
			className="h-full p-0 overflow-hidden"
		>
			<div className={`h-full overflow-y-auto px-6 py-4 ${agents.length === 0 ? "flex flex-col items-center justify-center" : ""}`}>
				{agents.length === 0 ? (
					<div className="w-full max-w-sm border border-border bg-card p-8 card-glow flex flex-col items-center justify-center text-center">
						<Monitor className="w-8 h-8 text-muted-foreground mb-3" />
						<p className="text-ui text-muted-foreground mb-1">
							No servers connected
						</p>
						<p className="text-label text-muted-foreground tracking-wider">
							Pair an agent to see your servers here
						</p>
					</div>
				) : (
					<div className="space-y-2">
						{agents.map((agent) => (
							<AgentListItem
								key={agent.id}
								agent={agent}
								onRemove={handleRemove}
							/>
						))}
					</div>
				)}
			</div>
		</PageLayout>
	);
}
