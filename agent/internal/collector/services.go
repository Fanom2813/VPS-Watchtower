package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const TypeSystemdServices = "metrics:services"

// ServiceInfo represents a systemd unit file.
type ServiceInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`     // "service", "timer", "socket"
	Enabled  bool   `json:"enabled"`  // has a symlink in wants/requires
	Modified int64  `json:"modified"` // file modification time (unix ms)
	Snippet  string `json:"snippet"`  // ExecStart line
}

// ServicesPayload is the payload for systemd services messages.
type ServicesPayload struct {
	Services  []ServiceInfo `json:"services"`
	Total     int           `json:"total"`
	Timestamp int64         `json:"timestamp"`
}

var unitDirs = []string{
	"/etc/systemd/system",
	"/usr/lib/systemd/system",
	"/lib/systemd/system",
}

var enabledDirs = []string{
	"/etc/systemd/system/multi-user.target.wants",
	"/etc/systemd/system/default.target.wants",
}

// SystemdCollector gathers systemd service unit files.
func SystemdCollector(interval time.Duration) *Collector {
	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeSystemdServices, ServicesPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		enabled := buildEnabledSet()
		services := scanUnitFiles(enabled)

		return TypeSystemdServices, ServicesPayload{
			Services:  services,
			Total:     len(services),
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

func buildEnabledSet() map[string]bool {
	set := make(map[string]bool)
	for _, dir := range enabledDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			set[e.Name()] = true
		}
	}
	return set
}

func scanUnitFiles(enabled map[string]bool) []ServiceInfo {
	seen := make(map[string]bool)
	var services []ServiceInfo

	for _, dir := range unitDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if seen[name] {
				continue
			}

			ext := filepath.Ext(name)
			if ext != ".service" && ext != ".timer" && ext != ".socket" {
				continue
			}

			// Skip template units
			if strings.Contains(name, "@") {
				continue
			}

			seen[name] = true
			path := filepath.Join(dir, name)

			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			svc := ServiceInfo{
				Name:     name,
				Path:     path,
				Type:     strings.TrimPrefix(ext, "."),
				Enabled:  enabled[name],
				Modified: info.ModTime().UnixMilli(),
				Snippet:  readExecStart(path),
			}

			services = append(services, svc)
		}
	}

	return services
}

func readExecStart(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ExecStart=") {
			return strings.TrimPrefix(line, "ExecStart=")
		}
	}
	return ""
}
