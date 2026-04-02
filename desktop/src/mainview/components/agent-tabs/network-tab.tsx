import type { NetworkPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";
import { Globe, Server } from "lucide-react";

interface NetworkTabProps {
  data?: NetworkPayload;
}

export function NetworkTab({ data }: NetworkTabProps) {
  const hasConnections = data?.connections && data.connections.length > 0;
  const hasListening = data?.listening && data.listening.length > 0;

  if (!hasConnections && !hasListening) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Globe className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Network Activity</EmptyTitle>
          <EmptyDescription>
            No active connections or listening ports detected.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div className="border border-border bg-card">
        <div className="px-4 py-3 border-b border-border">
          <h3 className="text-label font-semibold uppercase tracking-wider text-foreground">
            Active Connections
          </h3>
        </div>
        <div className="max-h-[300px] overflow-auto">
          {hasConnections ? (
            <table className="w-full text-sm">
              <tbody className="divide-y divide-border">
                {data.connections.map((conn, i) => (
                  <tr key={i} className="hover:bg-muted/30">
                    <td className="px-4 py-2 font-mono text-xs">
                      {conn.localAddr}:{conn.localPort}
                    </td>
                    <td className="px-4 py-2 text-muted-foreground">→</td>
                    <td className="px-4 py-2 font-mono text-xs">
                      {conn.remoteAddr}:{conn.remotePort}
                    </td>
                    <td className="px-4 py-2">
                      <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-smui-aurora-green/20 text-smui-aurora-green">
                        {conn.state}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <Empty className="py-12">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Globe className="w-4 h-4" />
                </EmptyMedia>
                <EmptyTitle>No Active Connections</EmptyTitle>
              </EmptyHeader>
            </Empty>
          )}
        </div>
      </div>

      <div className="border border-border bg-card">
        <div className="px-4 py-3 border-b border-border">
          <h3 className="text-label font-semibold uppercase tracking-wider text-foreground">
            Listening Ports
          </h3>
        </div>
        <div className="max-h-[300px] overflow-auto">
          {hasListening ? (
            <table className="w-full text-sm">
              <tbody className="divide-y divide-border">
                {data.listening.map((conn, i) => (
                  <tr key={i} className="hover:bg-muted/30">
                    <td className="px-4 py-2 font-mono text-xs">{conn.localAddr}</td>
                    <td className="px-4 py-2 font-mono text-xs font-medium">:{conn.localPort}</td>
                    <td className="px-4 py-2">
                      <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-primary/20 text-primary">
                        LISTEN
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <Empty className="py-12">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Server className="w-4 h-4" />
                </EmptyMedia>
                <EmptyTitle>No Listening Ports</EmptyTitle>
              </EmptyHeader>
            </Empty>
          )}
        </div>
      </div>
    </div>
  );
}
