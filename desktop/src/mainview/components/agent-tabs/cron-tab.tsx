import { Clock } from "lucide-react";
import type { CronPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface CronTabProps {
  data?: CronPayload;
}

export function CronTab({ data }: CronTabProps) {
  if (!data?.jobs || data.jobs.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Clock className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No Cron Jobs</EmptyTitle>
          <EmptyDescription>
            No scheduled cron jobs found on this system.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <Clock className="w-4 h-4" />
          Cron Jobs
        </h3>
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Schedule</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Command</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">User</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Source</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.jobs.map((job, i) => (
              <tr key={i} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-mono text-xs">{job.schedule}</td>
                <td className="px-4 py-2 text-xs truncate max-w-[300px]">{job.command}</td>
                <td className="px-4 py-2 text-xs">{job.user}</td>
                <td className="px-4 py-2 text-xs text-muted-foreground truncate max-w-[200px]">{job.source}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
