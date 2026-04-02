import { Container } from "lucide-react";
import type { DockerPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface DockerTabProps {
  data?: DockerPayload;
}

export function DockerTab({ data }: DockerTabProps) {
  if (data?.available === false) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Container className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>Docker Not Available</EmptyTitle>
          <EmptyDescription>
            Docker is not running or not installed on this host.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  if (!data?.containers || data.containers.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Container className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Containers</EmptyTitle>
          <EmptyDescription>
            No Docker containers are running on this host.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <Container className="w-4 h-4" />
          Containers
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">ID</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Name</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Image</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">State</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.containers.map((container) => (
              <tr key={container.id} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-mono text-xs">{container.id}</td>
                <td className="px-4 py-2">{container.name}</td>
                <td className="px-4 py-2 text-muted-foreground">{container.image}</td>
                <td className="px-4 py-2">
                  <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium uppercase ${
                    container.state === "running"
                      ? "bg-smui-aurora-green/20 text-smui-aurora-green"
                      : "bg-muted text-muted-foreground"
                  }`}>
                    {container.state}
                  </span>
                </td>
                <td className="px-4 py-2 text-xs text-muted-foreground">{container.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
