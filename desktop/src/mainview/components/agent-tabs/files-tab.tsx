import { FileCheck } from "lucide-react";
import type { FileIntegrityPayload } from "@/stores/agents";
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";

interface FilesTabProps {
  data?: FileIntegrityPayload;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

export function FilesTab({ data }: FilesTabProps) {
  if (!data?.files || data.files.length === 0) {
    return (
      <Empty className="h-full min-h-[300px]">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <FileCheck className="w-4 h-4" />
          </EmptyMedia>
          <EmptyTitle>No File Data</EmptyTitle>
          <EmptyDescription>
            File integrity monitoring data is not available yet.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <div className="border border-border bg-card">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <h3 className="text-label font-semibold uppercase tracking-wider text-foreground flex items-center gap-2">
          <FileCheck className="w-4 h-4" />
          File Integrity Monitor
        </h3>
        {data?.changes ? (
          <span className="px-1.5 py-0.5 text-[10px] bg-destructive text-destructive-foreground rounded-full">
            {data.changes} changes
          </span>
        ) : null}
      </div>
      <div className="max-h-[400px] overflow-auto">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 sticky top-0">
            <tr>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Path</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Size</th>
              <th className="text-left px-4 py-2 text-xs font-medium text-muted-foreground">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {data.files.map((file, i) => (
              <tr key={i} className="hover:bg-muted/30">
                <td className="px-4 py-2 font-mono text-xs">{file.path}</td>
                <td className="px-4 py-2 text-xs">{formatBytes(file.size)}</td>
                <td className="px-4 py-2">
                  {file.missing ? (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-destructive/20 text-destructive">
                      MISSING
                    </span>
                  ) : file.changed ? (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-warning/20 text-warning">
                      CHANGED
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-smui-aurora-green/20 text-smui-aurora-green">
                      OK
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
