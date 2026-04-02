import { Eye, ShieldCheck } from "lucide-react";
import type { SensitiveAccessPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface SensitiveTabProps {
  data?: SensitiveAccessPayload;
}

export function SensitiveTab({ data }: SensitiveTabProps) {
  if (!data?.accesses || data.accesses.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ShieldCheck className="w-4 h-4 text-smui-aurora-green" />
          </EmptyMedia>
          <EmptyTitle className="text-smui-aurora-green">No Sensitive Access</EmptyTitle>
          <EmptyDescription>
            No processes are currently accessing sensitive files.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <Eye className="w-4 h-4" />
          Sensitive File Access
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Process</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">PID</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">File</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Reason</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.accesses.map((access, i) => (
              <tr key={i} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-medium text-sm">{access.process}</td>
                <td className="px-4 py-2 font-mono text-xs">{access.pid}</td>
                <td className="px-4 py-2 font-mono text-xs">{access.file}</td>
                <td className="px-4 py-2 text-xs text-muted-foreground">{access.reason}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
