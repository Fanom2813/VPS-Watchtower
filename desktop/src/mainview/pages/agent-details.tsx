import { useNavigate, useParams } from "react-router";
import { useAgentsStore } from "@/stores/agents";
import { PageLayout } from "@/components/layout/page-layout";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  LayoutDashboard,
  Cpu,
  Terminal,
  Container,
  AlertTriangle,
  Globe,
  Clock,
  Settings,
  FileCheck,
  Eye,
} from "lucide-react";
import {
  OverviewTab,
  ProcessesTab,
  NetworkTab,
  AuthTab,
  DockerTab,
  IntrusionTab,
  OutboundTab,
  CronTab,
  ServicesTab,
  FilesTab,
  SensitiveTab,
} from "@/components/agent-tabs";

export function AgentDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { agents } = useAgentsStore();
  const agent = agents.find((a) => a.id === id);

  if (!agent) {
    return (
      <PageLayout
        title="Agent not found"
        onBack={() => navigate("/overview")}
        className="p-6"
      >
        <div className="flex items-center justify-center h-64 border border-border bg-card">
          <p className="text-muted-foreground">The requested agent could not be found.</p>
        </div>
      </PageLayout>
    );
  }

  const displayName = agent.label || agent.hostname || agent.id.slice(0, 12);
  const data = agent.collectorData;

  return (
    <PageLayout
      title={displayName}
      subtitle={`${agent.distro || agent.os} · ${agent.arch}`}
      onBack={() => navigate("/overview")}
      className="flex flex-col h-full p-0"
    >
      {/* Tabs fill available space */}
      <Tabs defaultValue="overview" className="flex flex-col flex-1 min-h-0">
        {/* Scrollable content area */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          <TabsContent value="overview" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <OverviewTab agent={agent} data={data} />
          </TabsContent>

          <TabsContent value="processes" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <ProcessesTab data={data.processes} />
          </TabsContent>

          <TabsContent value="network" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <NetworkTab data={data.network} />
          </TabsContent>

          <TabsContent value="auth" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <AuthTab data={data.authlog} />
          </TabsContent>

          <TabsContent value="docker" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <DockerTab data={data.docker} />
          </TabsContent>

          <TabsContent value="intrusion" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <IntrusionTab data={data.intrusion} />
          </TabsContent>

          <TabsContent value="outbound" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <OutboundTab data={data.outbound} />
          </TabsContent>

          <TabsContent value="cron" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <CronTab data={data.cron} />
          </TabsContent>

          <TabsContent value="services" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <ServicesTab data={data.services} />
          </TabsContent>

          <TabsContent value="files" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <FilesTab data={data.fileIntegrity} />
          </TabsContent>

          <TabsContent value="sensitive" className="mt-0 h-full flex flex-col [&[data-state='inactive']]:hidden">
            <SensitiveTab data={data.sensitiveAccess} />
          </TabsContent>
        </div>

        {/* Bottom Tab Bar - in flow, not fixed */}
        <TabsList className="h-8 w-full rounded-none border-t border-border bg-card/95 backdrop-blur-sm px-4 gap-2 shrink-0 justify-center">
          <TabsTrigger value="overview" className="h-6 w-6 p-0" title="Overview">
            <LayoutDashboard className="w-4 h-4" />
          </TabsTrigger>

          <TabsTrigger value="processes" className="h-6 w-6 p-0 relative" title="Processes">
            <Cpu className="w-4 h-4" />
            {data.processes?.unknown ? (
              <span className="absolute top-0.5 right-0.5 w-1.5 h-1.5 bg-warning rounded-full" />
            ) : null}
          </TabsTrigger>

          <TabsTrigger value="network" className="h-6 w-6 p-0" title="Network">
            <Globe className="w-4 h-4" />
          </TabsTrigger>

          <TabsTrigger value="auth" className="h-6 w-6 p-0" title="Auth Log">
            <Terminal className="w-4 h-4" />
          </TabsTrigger>

          <TabsTrigger value="docker" className="h-6 w-6 p-0 relative" title="Docker">
            <Container className="w-4 h-4" />
            {data.docker?.total ? (
              <span className="absolute -top-0.5 -right-0.5 px-1 py-0 text-[6px] bg-primary text-primary-foreground rounded-full min-w-[10px]">
                {data.docker.total}
              </span>
            ) : null}
          </TabsTrigger>

          <TabsTrigger value="intrusion" className="h-6 w-6 p-0 relative" title="Intrusion Alerts">
            <AlertTriangle className="w-4 h-4" />
            {data.intrusion?.alerts && data.intrusion.alerts.length > 0 && (
              <span className="absolute -top-0.5 -right-0.5 px-1 py-0 text-[6px] bg-destructive text-destructive-foreground rounded-full font-medium min-w-[10px]">
                {data.intrusion.alerts.length}
              </span>
            )}
          </TabsTrigger>

          <TabsTrigger value="outbound" className="h-6 w-6 p-0 relative" title="Outbound">
            <Globe className="w-4 h-4 rotate-45" />
            {data.outbound?.newDestinations ? (
              <span className="absolute top-0.5 right-0.5 w-1.5 h-1.5 bg-warning rounded-full" />
            ) : null}
          </TabsTrigger>

          <TabsTrigger value="cron" className="h-6 w-6 p-0" title="Cron Jobs">
            <Clock className="w-4 h-4" />
          </TabsTrigger>

          <TabsTrigger value="services" className="h-6 w-6 p-0" title="Services">
            <Settings className="w-4 h-4" />
          </TabsTrigger>

          <TabsTrigger value="files" className="h-6 w-6 p-0 relative" title="File Integrity">
            <FileCheck className="w-4 h-4" />
            {data.fileIntegrity?.changes && data.fileIntegrity.changes > 0 && (
              <span className="absolute -top-0.5 -right-0.5 px-1 py-0 text-[6px] bg-destructive text-destructive-foreground rounded-full font-medium min-w-[10px]">
                {data.fileIntegrity.changes}
              </span>
            )}
          </TabsTrigger>

          <TabsTrigger value="sensitive" className="h-6 w-6 p-0 relative" title="Sensitive Files">
            <Eye className="w-4 h-4" />
            {data.sensitiveAccess?.accesses && data.sensitiveAccess.accesses.length > 0 && (
              <span className="absolute -top-0.5 -right-0.5 px-1 py-0 text-[6px] bg-warning text-warning-foreground rounded-full font-medium min-w-[10px]">
                {data.sensitiveAccess.total}
              </span>
            )}
          </TabsTrigger>
        </TabsList>
      </Tabs>
    </PageLayout>
  );
}
