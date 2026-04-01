import { Trash2, Activity, Share2 } from "lucide-react";
import { useNavigate } from "react-router";
import type { IconType } from "@icons-pack/react-simple-icons";
import {
	SiLinux,
	SiUbuntu,
	SiDebian,
	SiCentos,
	SiRedhat,
	SiFedora,
	SiAlpinelinux,
	SiArchlinux,
	SiApple,
	SiDocker,
	SiKubernetes,
} from "@icons-pack/react-simple-icons";
import { Button } from "@/components/ui/button";
import type { Agent } from "@/stores/agents";

import { WindowsIcon } from "./icons/windows-icon";

interface AgentListItemProps {
	agent: Agent;
	onRemove: (id: string) => void;
}

function getOSIcon(distro?: string, os?: string): IconType | typeof WindowsIcon {
	const name = (distro || os || "").toLowerCase();

	if (name.includes("windows")) return WindowsIcon;
	if (name.includes("ubuntu")) return SiUbuntu;
	if (name.includes("debian")) return SiDebian;
	if (name.includes("centos")) return SiCentos;
	if (name.includes("redhat") || name.includes("rhel") || name.includes("red hat"))
		return SiRedhat;
	if (name.includes("fedora")) return SiFedora;
	if (name.includes("alpine")) return SiAlpinelinux;
	if (name.includes("arch")) return SiArchlinux;
	if (name.includes("darwin") || name.includes("macos") || name.includes("os x"))
		return SiApple;
	if (name.includes("docker")) return SiDocker;
	if (name.includes("kubernetes") || name.includes("k8s")) return SiKubernetes;

	// Default to generic Linux
	return SiLinux;
}

export function AgentListItem({ agent, onRemove }: AgentListItemProps) {
	const navigate = useNavigate();
	const OSIcon = getOSIcon(agent.distro, agent.os);
	const displayName = agent.label || agent.hostname || agent.id.slice(0, 12);

	return (
		<div
			onClick={() => navigate(`/agents/${agent.id}`)}
			className="border border-border bg-card px-4 py-3 card-glow flex items-center justify-between cursor-pointer hover:border-primary/50 transition-colors group"
		>
			<div className="flex items-center gap-3">
				<OSIcon className="w-8 h-8 text-muted-foreground group-hover:text-primary transition-colors" />
				<div className="flex flex-col gap-1">
					<div className="flex items-center gap-2">
						<p className="text-ui text-foreground font-medium">
							{displayName}
						</p>
						<span
							className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium uppercase tracking-wider ${
								agent.status === "online"
									? "bg-smui-aurora-green/20 text-smui-aurora-green"
									: "bg-muted text-muted-foreground"
							}`}
						>
							{agent.status}
						</span>
					</div>
					<div className="flex items-center gap-2 text-label text-muted-foreground tracking-wider">
						<span>
							{agent.os && `${agent.distro || agent.os} · `}
							{agent.arch}
						</span>
						{agent.status === "online" && (
							<>
								<span className="text-border">|</span>
								<span className="flex items-center gap-1">
									<Activity className="w-3 h-3" />
									--% CPU · --% MEM
								</span>
							</>
						)}
					</div>
				</div>
			</div>
			<div className="flex items-center gap-1">
				<Button
					variant="ghost"
					size="icon-sm"
					onClick={(e) => {
						e.stopPropagation();
						// Share functionality
					}}
					className="text-muted-foreground hover:text-foreground"
				>
					<Share2 className="w-3.5 h-3.5" />
				</Button>
				<Button
					variant="ghost"
					size="icon-sm"
					onClick={(e) => {
						e.stopPropagation();
						onRemove(agent.id);
					}}
					className="text-muted-foreground hover:text-destructive"
				>
					<Trash2 className="w-3.5 h-3.5" />
				</Button>
			</div>
		</div>
	);
}
