import {
  Activity,
  Cpu,
  HardDrive,
  Network,
  Shield,
} from "lucide-react";
import type { Agent, AgentCollectorData } from "@/stores/agents";

interface OverviewTabProps {
  agent: Agent;
  data: AgentCollectorData;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  if (days > 0) return `${days}d ${hours}h ${mins}m`;
  if (hours > 0) return `${hours}h ${mins}m`;
  return `${mins}m`;
}

export function OverviewTab({ agent, data }: OverviewTabProps) {
  const system = data.system;

  return (
    <div className="space-y-4">
      {/* Overview Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="border border-border bg-card p-4 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-label text-muted-foreground uppercase tracking-wider">CPU Usage</span>
            <Cpu className="w-4 h-4 text-primary" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold">
              {system?.cpuPercent !== undefined ? `${system.cpuPercent.toFixed(1)}%` : "--%"}
            </span>
            <span className="text-xs text-muted-foreground font-mono">/ 100%</span>
          </div>
          {system?.loadAvg && (
            <div className="text-xs text-muted-foreground font-mono">
              Load: {system.loadAvg[0].toFixed(2)} {system.loadAvg[1].toFixed(2)} {system.loadAvg[2].toFixed(2)}
            </div>
          )}
        </div>

        <div className="border border-border bg-card p-4 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-label text-muted-foreground uppercase tracking-wider">Memory</span>
            <Activity className="w-4 h-4 text-smui-aurora-green" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold">
              {system?.memPercent !== undefined ? `${system.memPercent.toFixed(1)}%` : "--%"}
            </span>
            <span className="text-xs text-muted-foreground font-mono">
              {system ? `${formatBytes(system.memUsed)} / ${formatBytes(system.memTotal)}` : "0 / 0"}
            </span>
          </div>
        </div>

        <div className="border border-border bg-card p-4 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-label text-muted-foreground uppercase tracking-wider">Storage</span>
            <HardDrive className="w-4 h-4 text-smui-aurora-yellow" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold">
              {system?.diskPercent !== undefined ? `${system.diskPercent.toFixed(1)}%` : "--%"}
            </span>
            <span className="text-xs text-muted-foreground font-mono">
              {system ? `${formatBytes(system.diskUsed)} / ${formatBytes(system.diskTotal)}` : "0 / 0"}
            </span>
          </div>
        </div>

        <div className="border border-border bg-card p-4 space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-label text-muted-foreground uppercase tracking-wider">Network</span>
            <Network className="w-4 h-4 text-smui-aurora-blue" />
          </div>
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold">
              {system?.netRxBytes !== undefined ? `${formatBytes(system.netRxBytes)}/s` : "--"}
            </span>
            <span className="text-xs text-muted-foreground font-mono">RX / TX</span>
          </div>
          {system?.uptime && (
            <div className="text-xs text-muted-foreground">
              Uptime: {formatUptime(system.uptime)}
            </div>
          )}
        </div>
      </div>

      {/* Agent Info */}
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
            {data.tamper?.modified !== undefined && (
              <div className="flex justify-between text-ui">
                <span className="text-muted-foreground">Binary Integrity</span>
                <span className={data.tamper.modified ? "text-destructive" : "text-smui-aurora-green"}>
                  {data.tamper.modified ? "MODIFIED" : "OK"}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
