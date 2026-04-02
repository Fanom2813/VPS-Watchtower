import { AlertTriangle, ShieldCheck } from "lucide-react";
import type { IntrusionPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface IntrusionTabProps {
  data?: IntrusionPayload;
}

export function IntrusionTab({ data }: IntrusionTabProps) {
  if (!data?.alerts || data.alerts.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ShieldCheck className="w-4 h-4 text-smui-aurora-green" />
          </EmptyMedia>
          <EmptyTitle className="text-smui-aurora-green">All Clear</EmptyTitle>
          <EmptyDescription>
            No security alerts detected. Your system looks secure.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <AlertTriangle className="w-4 h-4" />
          Security Alerts
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <div className="divide-y divide-border">
          {data.alerts.map((alert, i) => (
            <div key={i} className="px-4 py-3 hover:bg-muted/30">
              <div className="flex items-center gap-2 mb-1">
                <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium uppercase ${
                  alert.severity === "critical"
                    ? "bg-destructive text-destructive-foreground"
                    : alert.severity === "high"
                    ? "bg-destructive/80 text-destructive-foreground"
                    : alert.severity === "medium"
                    ? "bg-warning text-warning-foreground"
                    : "bg-muted text-muted-foreground"
                }`}>
                  {alert.severity}
                </span>
                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-muted text-muted-foreground uppercase">
                  {alert.type.replace("_", " ")}
                </span>
              </div>
              <div className="text-sm font-medium">{alert.detail}</div>
              <div className="flex items-center gap-4 mt-1 text-xs text-muted-foreground">
                <span>Source: {alert.source}</span>
                {alert.count > 1 && <span>Count: {alert.count}</span>}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
