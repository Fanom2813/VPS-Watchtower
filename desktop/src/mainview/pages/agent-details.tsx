import { useNavigate, useParams } from "react-router";
import { useAgentsStore } from "@/stores/agents";
import { PageLayout } from "@/components/layout/page-layout";
import { Activity, Cpu, HardDrive, Network, Shield } from "lucide-react";

export function AgentDetailsPage() {
	const { id } = useParams();
	const navigate = useNavigate();
	const { agents } = useAgentsStore();
	const agent = agents.find((a) => a.id === id);

	if (!agent) {
		return (
			<PageLayout
				title="Agent not found"
				onBack={() => navigate("/overview")}
			>
				<div className="flex items-center justify-center h-64 border border-border bg-card">
					<p className="text-muted-foreground">The requested agent could not be found.</p>
				</div>
			</PageLayout>
		);
	}

	const displayName = agent.label || agent.hostname || agent.id.slice(0, 12);

	return (
		<PageLayout
			title={displayName}
			subtitle={`${agent.distro || agent.os} · ${agent.arch}`}
			onBack={() => navigate("/overview")}
		>
			<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
				{/* Stats Cards - Placeholders for now */}
				<div className="border border-border bg-card p-4 space-y-2">
					<div className="flex items-center justify-between">
						<span className="text-label text-muted-foreground uppercase tracking-wider">CPU Usage</span>
						<Cpu className="w-4 h-4 text-primary" />
					</div>
					<div className="flex items-baseline gap-2">
						<span className="text-2xl font-semibold">--%</span>
						<span className="text-xs text-muted-foreground font-mono">/ 100%</span>
					</div>
				</div>

				<div className="border border-border bg-card p-4 space-y-2">
					<div className="flex items-center justify-between">
						<span className="text-label text-muted-foreground uppercase tracking-wider">Memory</span>
						<Activity className="w-4 h-4 text-smui-aurora-green" />
					</div>
					<div className="flex items-baseline gap-2">
						<span className="text-2xl font-semibold">--%</span>
						<span className="text-xs text-muted-foreground font-mono">0.0 / 0.0 GB</span>
					</div>
				</div>

				<div className="border border-border bg-card p-4 space-y-2">
					<div className="flex items-center justify-between">
						<span className="text-label text-muted-foreground uppercase tracking-wider">Storage</span>
						<HardDrive className="w-4 h-4 text-smui-aurora-yellow" />
					</div>
					<div className="flex items-baseline gap-2">
						<span className="text-2xl font-semibold">--%</span>
						<span className="text-xs text-muted-foreground font-mono">0.0 / 0.0 GB</span>
					</div>
				</div>

				<div className="border border-border bg-card p-4 space-y-2">
					<div className="flex items-center justify-between">
						<span className="text-label text-muted-foreground uppercase tracking-wider">Network</span>
						<Network className="w-4 h-4 text-smui-aurora-blue" />
					</div>
					<div className="flex items-baseline gap-2">
						<span className="text-2xl font-semibold">-- Mb/s</span>
						<span className="text-xs text-muted-foreground font-mono">UP / DOWN</span>
					</div>
				</div>
			</div>

			<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<div className="border border-border bg-card">
					<div className="px-4 py-3 border-b border-border flex items-center justify-between">
						<h3 className="text-label font-semibold uppercase tracking-wider text-foreground">
							System Information
						</h3>
					</div>
					<div className="p-4 space-y-3">
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Hostname</span>
							<span className="font-mono">{agent.hostname}</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">OS</span>
							<span>{agent.os}</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Distro</span>
							<span>{agent.distro}</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Architecture</span>
							<span className="font-mono">{agent.arch}</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Agent Version</span>
							<span className="font-mono">v{agent.agentVersion}</span>
						</div>
					</div>
				</div>

				<div className="border border-border bg-card">
					<div className="px-4 py-3 border-b border-border flex items-center justify-between">
						<h3 className="text-label font-semibold uppercase tracking-wider text-foreground">
							Security & Connectivity
						</h3>
						<Shield className="w-4 h-4 text-muted-foreground" />
					</div>
					<div className="p-4 space-y-3">
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Status</span>
							<span className={`uppercase font-semibold tracking-wider ${
								agent.status === "online" ? "text-smui-aurora-green" : "text-muted-foreground"
							}`}>
								{agent.status}
							</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Endpoint</span>
							<span className="font-mono text-xs">{agent.url}</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Paired At</span>
							<span>{new Date(agent.pairedAt).toLocaleString()}</span>
						</div>
						<div className="flex justify-between text-ui">
							<span className="text-muted-foreground">Last Seen</span>
							<span>{agent.lastSeen > 0 ? new Date(agent.lastSeen).toLocaleString() : "Never"}</span>
						</div>
					</div>
				</div>
			</div>
		</PageLayout>
	);
}
