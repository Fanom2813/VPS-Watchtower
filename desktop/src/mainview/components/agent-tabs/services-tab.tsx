import { Settings } from "lucide-react";
import type { ServicesPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface ServicesTabProps {
  data?: ServicesPayload;
}

export function ServicesTab({ data }: ServicesTabProps) {
  if (!data?.services || data.services.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Settings className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Services</EmptyTitle>
          <EmptyDescription>
            No systemd services found on this system.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <Settings className="w-4 h-4" />
          Systemd Services
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Name</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Type</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Status</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">ExecStart</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.services.map((svc, i) => (
              <tr key={i} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-medium text-sm">{svc.name}</td>
                <td className="px-4 py-2">
                  <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-muted text-muted-foreground uppercase">
                    {svc.type}
                  </span>
                </td>
                <td className="px-4 py-2">
                  {svc.enabled ? (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-smui-aurora-green/20 text-smui-aurora-green">
                      enabled
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-muted text-muted-foreground">
                      disabled
                    </span>
                  )}
                </td>
                <td className="px-4 py-2 text-xs text-muted-foreground truncate max-w-[300px]">{svc.snippet || "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
