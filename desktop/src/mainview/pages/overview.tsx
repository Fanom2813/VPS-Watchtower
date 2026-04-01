import { useEffect } from "react";
import { useNavigate } from "react-router";
import { Monitor, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAgentsStore } from "@/stores/agents";

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

	return (
		<div className="px-6 py-6 space-y-4">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-heading font-semibold tracking-[1.5px] uppercase text-foreground">
						Servers
					</h1>
					<p className="text-ui text-muted-foreground mt-1">
						Your connected VPS instances
					</p>
				</div>
				<Button
					variant="outline"
					className="uppercase tracking-wider text-label"
					onClick={() => navigate("/add")}
				>
					<Plus className="w-3.5 h-3.5 mr-2" />
					Add Server
				</Button>
			</div>

			{agents.length === 0 ? (
				<div className="border border-border bg-card p-8 card-glow flex flex-col items-center justify-center text-center">
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
						<div
							key={agent.id}
							className="border border-border bg-card px-4 py-3 card-glow flex items-center justify-between"
						>
							<div className="flex items-center gap-3">
								<div
									className={`w-2 h-2 rounded-full ${
										agent.status === "online"
											? "bg-smui-aurora-green"
											: "bg-muted-foreground/40"
									}`}
								/>
								<div>
									<p className="text-ui text-foreground font-medium">
										{agent.hostname || agent.id.slice(0, 12)}
									</p>
									<p className="text-label text-muted-foreground tracking-wider">
										{agent.os && `${agent.distro || agent.os} · `}
										{agent.arch}
										{agent.agentVersion && ` · ${agent.agentVersion}`}
									</p>
								</div>
							</div>
							<Button
								variant="ghost"
								size="icon-xs"
								onClick={() => handleRemove(agent.id)}
								className="text-muted-foreground hover:text-destructive"
							>
								<Trash2 className="w-3.5 h-3.5" />
							</Button>
						</div>
					))}
				</div>
			)}
		</div>
	);
}
