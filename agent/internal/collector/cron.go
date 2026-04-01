package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const TypeCronJobs = "metrics:cron"

// CronEntry represents a single cron job.
type CronEntry struct {
	Source   string `json:"source"`   // file path or "user:username"
	Schedule string `json:"schedule"` // cron schedule expression
	Command  string `json:"command"`
	User     string `json:"user"`
}

// CronPayload is the payload for cron job messages.
type CronPayload struct {
	Jobs      []CronEntry `json:"jobs"`
	Total     int         `json:"total"`
	Timestamp int64       `json:"timestamp"`
}

// CronCollector gathers all cron jobs from system and user crontabs.
func CronCollector(interval time.Duration) *Collector {
	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeCronJobs, CronPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		var jobs []CronEntry

		// System crontab
		jobs = append(jobs, parseCrontab("/etc/crontab", "root", true)...)

		// /etc/cron.d/*
		if entries, err := os.ReadDir("/etc/cron.d"); err == nil {
			for _, e := range entries {
				if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
					continue
				}
				path := filepath.Join("/etc/cron.d", e.Name())
				jobs = append(jobs, parseCrontab(path, "", true)...)
			}
		}

		// User crontabs in /var/spool/cron/crontabs/ (Debian/Ubuntu)
		// or /var/spool/cron/ (RHEL/CentOS)
		for _, dir := range []string{"/var/spool/cron/crontabs", "/var/spool/cron"} {
			if entries, err := os.ReadDir(dir); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					path := filepath.Join(dir, e.Name())
					jobs = append(jobs, parseCrontab(path, e.Name(), false)...)
				}
			}
		}

		return TypeCronJobs, CronPayload{
			Jobs:      jobs,
			Total:     len(jobs),
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

// parseCrontab reads a crontab file and extracts job entries.
// If systemFormat is true, the 6th field is the user (like /etc/crontab).
func parseCrontab(path string, defaultUser string, systemFormat bool) []CronEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entries []CronEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		// Skip variable assignments (NAME=value)
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "*") &&
			!strings.HasPrefix(line, "@") && (len(line) < 2 || line[1] < '0' || line[1] > '9') {
			if idx := strings.IndexByte(line, '='); idx > 0 && !strings.ContainsAny(line[:idx], " \t") {
				continue
			}
		}

		entry := parseCronLine(line, path, defaultUser, systemFormat)
		if entry.Command != "" {
			entries = append(entries, entry)
		}
	}

	return entries
}

func parseCronLine(line, source, defaultUser string, systemFormat bool) CronEntry {
	// Handle @reboot, @daily, etc.
	if strings.HasPrefix(line, "@") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return CronEntry{}
		}
		schedule := fields[0]
		user := defaultUser
		cmdStart := 1
		if systemFormat && len(fields) >= 3 {
			user = fields[1]
			cmdStart = 2
		}
		return CronEntry{
			Source:   source,
			Schedule: schedule,
			Command:  strings.Join(fields[cmdStart:], " "),
			User:     user,
		}
	}

	// Standard 5-field schedule
	fields := strings.Fields(line)
	minFields := 6
	if systemFormat {
		minFields = 7 // 5 schedule + user + command
	}
	if len(fields) < minFields {
		return CronEntry{}
	}

	schedule := strings.Join(fields[:5], " ")
	user := defaultUser
	cmdStart := 5
	if systemFormat {
		user = fields[5]
		cmdStart = 6
	}

	return CronEntry{
		Source:   source,
		Schedule: schedule,
		Command:  strings.Join(fields[cmdStart:], " "),
		User:     user,
	}
}
