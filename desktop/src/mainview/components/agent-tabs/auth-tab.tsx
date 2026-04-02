import { Terminal } from "lucide-react";
import type { AuthLogPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface AuthTabProps {
  data?: AuthLogPayload;
}

export function AuthTab({ data }: AuthTabProps) {
  if (!data?.entries || data.entries.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Terminal className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Auth Events</EmptyTitle>
          <EmptyDescription>
            No authentication events have been recorded yet.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <Terminal className="w-4 h-4" />
          Authentication Events
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <div className="divide-y divide-border">
          {data.entries.map((entry, i) => (
            <div key={i} className="px-4 py-3 hover:bg-muted/30">
              <div className="flex items-center gap-2 mb-1">
                <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium uppercase ${
                  entry.type === "login_success"
                    ? "bg-smui-aurora-green/20 text-smui-aurora-green"
                    : entry.type === "login_failed"
                    ? "bg-destructive/20 text-destructive"
                    : "bg-muted text-muted-foreground"
                }`}>
                  {entry.type.replace("_", " ")}
                </span>
                <span className="text-xs text-muted-foreground">{entry.timestamp}</span>
              </div>
              <div className="flex items-center gap-4 text-sm">
                <span><span className="text-muted-foreground">User:</span> {entry.user}</span>
                <span><span className="text-muted-foreground">From:</span> {entry.source}</span>
                {entry.method && <span><span className="text-muted-foreground">Method:</span> {entry.method}</span>}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
