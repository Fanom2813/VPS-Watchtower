import { useNavigate } from "react-router";
import { PageLayout } from "@/components/layout/page-layout";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";
import { rpc } from "@/lib/rpc";
import {
  ArrowRight,
  Download,
  RefreshCw,
  CheckCircle,
  Power,
  Minimize2,
  Laptop,
  ExternalLink,
} from "lucide-react";

interface ToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label: string;
  description?: string;
  icon?: React.ReactNode;
}

function ToggleSetting({ checked, onChange, label, description, icon }: ToggleProps) {
  return (
    <div className="flex items-start justify-between py-3 px-4">
      <div className="flex items-start gap-3">
        {icon && (
          <div className="mt-0.5 text-muted-foreground">
            {icon}
          </div>
        )}
        <div>
          <div className="text-sm font-medium text-foreground">{label}</div>
          {description && (
            <div className="text-xs text-muted-foreground mt-0.5">{description}</div>
          )}
        </div>
      </div>
      <button
        onClick={() => onChange(!checked)}
        className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
          checked ? "bg-primary" : "bg-muted"
        }`}
      >
        <span
          className={`inline-block h-3.5 w-3.5 transform rounded-full bg-background transition-transform ${
            checked ? "translate-x-5" : "translate-x-1"
          }`}
        />
      </button>
    </div>
  );
}

export function SettingsPage() {
  const navigate = useNavigate();
  const [settings, setSettings] = useState({
    autoStart: false,
    startInTray: false,
    minimizeToTray: false,
    autoUpdate: false,
  });

  const [updateInfo, setUpdateInfo] = useState<{
    available: boolean;
    version?: string;
    currentVersion?: string;
    checking: boolean;
    downloading: boolean;
    ready: boolean;
    error?: string;
  }>({
    available: false,
    checking: false,
    downloading: false,
    ready: false,
  });

  useEffect(() => {
    // Load settings from backend
    rpc.request.getSettings({}).then((s) => {
      setSettings(s);
    });
  }, []);

  const updateSetting = (key: keyof typeof settings, value: boolean) => {
    const newSettings = { ...settings, [key]: value };
    setSettings(newSettings);
    rpc.request.setSettings({ settings: newSettings });
  };

  const handleCheckUpdate = async () => {
    setUpdateInfo((prev) => ({ ...prev, checking: true, error: undefined }));
    try {
      const result = await rpc.request.checkForUpdate({});
      setUpdateInfo((prev) => ({
        ...prev,
        available: result.available,
        version: result.version,
        currentVersion: result.currentVersion,
        checking: false,
        error: result.error,
      }));
    } catch (e) {
      setUpdateInfo((prev) => ({
        ...prev,
        checking: false,
        error: "Failed to check for updates",
      }));
    }
  };

  const handleDownload = async () => {
    setUpdateInfo((prev) => ({ ...prev, downloading: true, error: undefined }));
    try {
      const result = await rpc.request.downloadUpdate({});
      setUpdateInfo((prev) => ({
        ...prev,
        downloading: false,
        ready: result.ready ?? false,
        error: result.error,
      }));
    } catch (e) {
      setUpdateInfo((prev) => ({
        ...prev,
        downloading: false,
        error: "Failed to download update",
      }));
    }
  };

  const handleApplyUpdate = () => {
    rpc.request.applyUpdate({});
  };

  const openGitHub = () => {
    rpc.request.openExternal({ url: "https://github.com/fanom2813/eyes-on-vps" });
  };

  return (
    <PageLayout
      title="Settings"
      onBack={() => navigate("/overview")}
      className="h-full p-0 overflow-hidden"
    >
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-6 pb-8">
        {/* General Settings */}
        <div>
          <h3 className="text-label font-semibold uppercase tracking-wider text-muted-foreground mb-3">
            General
          </h3>
          <div className="border border-border bg-card divide-y divide-border">
            <ToggleSetting
              checked={settings.autoStart}
              onChange={(v) => updateSetting("autoStart", v)}
              label="Start on system login"
              description="Automatically start the app when you log in"
              icon={<Power className="w-4 h-4" />}
            />
            <ToggleSetting
              checked={settings.startInTray}
              onChange={(v) => updateSetting("startInTray", v)}
              label="Start minimized to tray"
              description="Start the app in the system tray instead of showing the window"
              icon={<Minimize2 className="w-4 h-4" />}
            />
            <ToggleSetting
              checked={settings.minimizeToTray}
              onChange={(v) => updateSetting("minimizeToTray", v)}
              label="Minimize to tray"
              description="Keep the app running in the system tray when minimized"
              icon={<Laptop className="w-4 h-4" />}
            />
          </div>
        </div>

        {/* Updates */}
        <div>
          <h3 className="text-label font-semibold uppercase tracking-wider text-muted-foreground mb-3">
            Updates
          </h3>
          <div className="border border-border bg-card divide-y divide-border">
            {/* Auto Update Toggle */}
            <div className="flex items-start justify-between py-3 px-4">
              <div className="flex items-start gap-3">
                <div className="mt-0.5 text-muted-foreground">
                  <RefreshCw className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-sm font-medium text-foreground">Automatic Updates</div>
                  <div className="text-xs text-muted-foreground mt-0.5">
                    Automatically check and install updates in the background
                  </div>
                </div>
              </div>
              <button
                onClick={() => updateSetting("autoUpdate", !settings.autoUpdate)}
                className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                  settings.autoUpdate ? "bg-primary" : "bg-muted"
                }`}
              >
                <span
                  className={`inline-block h-3.5 w-3.5 transform rounded-full bg-background transition-transform ${
                    settings.autoUpdate ? "translate-x-5" : "translate-x-1"
                  }`}
                />
              </button>
            </div>

            {/* Update Status / Check */}
            <div className="p-4 space-y-4">
              {updateInfo.available ? (
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <CheckCircle className="w-4 h-4 text-smui-aurora-green" />
                    <span className="text-sm">
                      New version available: <strong>v{updateInfo.version}</strong>
                    </span>
                  </div>
                  {updateInfo.currentVersion && (
                    <p className="text-xs text-muted-foreground">
                      You are currently on v{updateInfo.currentVersion}
                    </p>
                  )}
                  {updateInfo.ready ? (
                    <Button
                      onClick={handleApplyUpdate}
                      className="w-full uppercase tracking-wider text-label"
                    >
                      Restart to Update
                      <ArrowRight className="w-3.5 h-3.5 ml-2" />
                    </Button>
                  ) : (
                    <Button
                      onClick={handleDownload}
                      disabled={updateInfo.downloading}
                      className="w-full uppercase tracking-wider text-label"
                    >
                      {updateInfo.downloading ? (
                        <>
                          <RefreshCw className="w-3.5 h-3.5 mr-2 animate-spin" />
                          Downloading...
                        </>
                      ) : (
                        <>
                          <Download className="w-3.5 h-3.5 mr-2" />
                          Download Update
                        </>
                      )}
                    </Button>
                  )}
                </div>
              ) : (
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm font-medium">Automatic Updates</div>
                    <div className="text-xs text-muted-foreground">
                      {updateInfo.checking
                        ? "Checking for updates..."
                        : updateInfo.error
                        ? "Update check failed"
                        : "You are on the latest version"}
                    </div>
                  </div>
                  <Button
                    onClick={handleCheckUpdate}
                    disabled={updateInfo.checking}
                    variant="outline"
                    size="sm"
                    className="uppercase tracking-wider text-label"
                  >
                    {updateInfo.checking ? (
                      <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                    ) : (
                      "Check Now"
                    )}
                  </Button>
                </div>
              )}
              {updateInfo.error && (
                <p className="text-xs text-destructive">{updateInfo.error}</p>
              )}
            </div>
          </div>
        </div>

        {/* About - Full width card */}
        <div className="border border-border bg-card">
          <div className="px-4 py-3 border-b border-border flex items-center justify-between">
            <h3 className="text-label font-semibold uppercase tracking-wider text-foreground">
              About
            </h3>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={openGitHub}
              title="Open GitHub"
            >
              <ExternalLink className="w-3.5 h-3.5" />
            </Button>
          </div>
          <div className="p-4">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
                <svg
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  className="text-primary"
                >
                  <circle cx="12" cy="12" r="3" />
                  <path
                    d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"
                    opacity="0.3"
                  />
                  <path d="M12 5v2M12 17v2M5 12h2M17 12h2" />
                </svg>
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-semibold text-foreground">Eyes on VPS</h4>
                <p className="text-xs text-muted-foreground mt-1">
                  Lightweight VPS monitoring with desktop dashboard and server agents.
                </p>
                <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                  <span>Version 1.0.0</span>
                  <span className="text-border">|</span>
                  <span>Built with Electrobun</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>
  );
}
