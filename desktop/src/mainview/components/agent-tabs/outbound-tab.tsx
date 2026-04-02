import { Globe } from "lucide-react";
import type { OutboundPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface OutboundTabProps {
  data?: OutboundPayload;
}

export function OutboundTab({ data }: OutboundTabProps) {
  if (!data?.connections || data.connections.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Globe className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Outbound Connections</EmptyTitle>
          <EmptyDescription>
            No outbound network connections are currently active.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground">
          Outbound Connections
        </h3>
        {data?.newDestinations ? (
          <span className="px-1.5 py-0.5 text-[10px] bg-warning/20 text-warning rounded-full">
            {data.newDestinations} new
          </span>
        ) : null}
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Destination</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Port</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Local Port</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.connections.map((conn, i) => (
              <tr key={i} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-mono text-xs">{conn.remoteAddr}</td>
                <td className="px-4 py-2 font-mono text-xs">{conn.remotePort}</td>
                <td className="px-4 py-2 font-mono text-xs">{conn.localPort}</td>
                <td className="px-4 py-2">
                  {conn.new ? (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-warning/20 text-warning">
                      NEW
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-muted text-muted-foreground">
                      established
                    </span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
