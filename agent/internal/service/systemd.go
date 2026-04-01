package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const (
	ServiceName = "eyes-agent"
	UnitPath    = "/etc/systemd/system/" + ServiceName + ".service"
	BinaryDir   = "/usr/local/bin"
	BinaryName  = "eyes-agent"
	ConfigDir   = "/etc/eyes-on-vps"
	ConfigFile  = "agent.json"
)

// BinaryPath returns the full path to the installed binary.
func BinaryPath() string {
	return filepath.Join(BinaryDir, BinaryName)
}

// ConfigPath returns the full path to the installed config.
func ConfigPath() string {
	return filepath.Join(ConfigDir, ConfigFile)
}

var unitTemplate = template.Must(template.New("unit").Parse(`[Unit]
Description=Eyes on VPS Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} run -config {{.ConfigPath}}
Restart=always
RestartSec=5
User={{.User}}
Group={{.Group}}

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths={{.ConfigDir}}
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`))

type unitData struct {
	BinaryPath string
	ConfigPath string
	ConfigDir  string
	User       string
	Group      string
}

// Install copies the binary, writes the systemd unit, and enables the service.
// Must be run as root.
func Install(currentBinary string) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("must be run as root (use sudo)")
	}

	if err := ensureUser(); err != nil {
		return fmt.Errorf("create service user: %w", err)
	}

	if err := installBinary(currentBinary); err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	if err := installConfig(); err != nil {
		return fmt.Errorf("install config: %w", err)
	}

	if err := writeUnit(); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}

	if err := enableService(); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}

	return nil
}

// Uninstall stops the service, removes the unit file, and cleans up.
// Does not remove config (user data).
func Uninstall() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("must be run as root (use sudo)")
	}

	// Stop and disable (ignore errors — may not be running).
	exec.Command("systemctl", "stop", ServiceName).Run()
	exec.Command("systemctl", "disable", ServiceName).Run()

	os.Remove(UnitPath)
	os.Remove(BinaryPath())

	exec.Command("systemctl", "daemon-reload").Run()

	return nil
}

// Start starts the systemd service.
func Start() error {
	return run("systemctl", "start", ServiceName)
}

// Status returns the service status output.
func Status() (string, error) {
	out, err := exec.Command("systemctl", "status", ServiceName).CombinedOutput()
	// systemctl status returns non-zero for inactive services, so just return the output.
	return string(out), err
}

func ensureUser() error {
	// Check if user already exists.
	if _, err := exec.Command("id", ServiceName).Output(); err == nil {
		return nil
	}

	return run("useradd", "--system", "--no-create-home", "--shell", "/usr/sbin/nologin", ServiceName)
}

func installBinary(src string) error {
	if err := os.MkdirAll(BinaryDir, 0o755); err != nil {
		return err
	}

	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(srcAbs)
	if err != nil {
		return fmt.Errorf("read source binary: %w", err)
	}

	if err := os.WriteFile(BinaryPath(), data, 0o755); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}

	return nil
}

func installConfig() error {
	if err := os.MkdirAll(ConfigDir, 0o700); err != nil {
		return err
	}

	// If a local agent.json exists, copy it to the system location.
	if _, err := os.Stat("agent.json"); err == nil {
		data, err := os.ReadFile("agent.json")
		if err != nil {
			return err
		}
		if err := os.WriteFile(ConfigPath(), data, 0o600); err != nil {
			return err
		}
	} else if _, err := os.Stat(ConfigPath()); os.IsNotExist(err) {
		return fmt.Errorf("no config found — run 'agent setup' first")
	}

	// Ensure the service user owns the config directory.
	return run("chown", "-R", ServiceName+":"+ServiceName, ConfigDir)
}

func writeUnit() error {
	f, err := os.Create(UnitPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return unitTemplate.Execute(f, unitData{
		BinaryPath: BinaryPath(),
		ConfigPath: ConfigPath(),
		ConfigDir:  ConfigDir,
		User:       ServiceName,
		Group:      ServiceName,
	})
}

func enableService() error {
	if err := run("systemctl", "daemon-reload"); err != nil {
		return err
	}
	return run("systemctl", "enable", ServiceName)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
