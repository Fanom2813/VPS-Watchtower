import { Cpu } from "lucide-react";
import type { ProcessList } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface ProcessesTabProps {
  data?: ProcessList;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

export function ProcessesTab({ data }: ProcessesTabProps) {
  if (!data?.processes || data.processes.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Cpu className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Processes</EmptyTitle>
          <EmptyDescription>
            No process data available from this agent yet.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <Cpu className="w-4 h-4" />
          Process List
          {data?.unknown ? (
            <span className="px-1.5 py-0.5 text-[10px] bg-warning/20 text-warning rounded-full">
              {data.unknown} unknown
            </span>
          ) : null}
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">PID</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Name</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">User</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Class</th>
              <th className="text-right px-4 py-2 text-xs font-medium text-muted-foreground">Memory</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.processes.slice(0, 100).map((proc) => (
              <tr key={proc.pid} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-mono text-xs">{proc.pid}</td>
                <td className="px-4 py-2">
                  <div className="flex flex-col">
                    <span className="font-medium">{proc.name}</span>
                    {proc.binPath && (
                      <span className="text-xs text-muted-foreground truncate max-w-[200px]">
                        {proc.binPath}
                      </span>
                    )}
                  </div>
                </td>
                <td className="px-4 py-2 text-xs">{proc.user}</td>
                <td className="px-4 py-2">
                  <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium uppercase ${
                    proc.class === "system"
                      ? "bg-smui-aurora-green/20 text-smui-aurora-green"
                      : proc.class === "kernel"
                      ? "bg-muted text-muted-foreground"
                      : "bg-warning/20 text-warning"
                  }`}>
                    {proc.class}
                  </span>
                </td>
                <td className="px-4 py-2 text-right font-mono text-xs">{formatBytes(proc.memRss)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
