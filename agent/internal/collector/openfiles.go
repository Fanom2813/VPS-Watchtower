package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const TypeSensitiveAccess = "metrics:sensitive_access"

// Sensitive file paths to monitor.
var sensitiveFiles = map[string]string{
	"/etc/shadow":           "password hashes",
	"/etc/passwd":           "user accounts",
	"/etc/sudoers":          "sudo configuration",
	"/etc/ssh/sshd_config":  "SSH server config",
	"/root/.ssh/authorized_keys": "root SSH keys",
	"/root/.bash_history":   "root shell history",
	"/root/.env":            "root environment secrets",
}

// SensitiveAccess represents a process accessing a sensitive file.
type SensitiveAccess struct {
	PID      int    `json:"pid"`
	Process  string `json:"process"`
	File     string `json:"file"`
	Reason   string `json:"reason"` // why the file is sensitive
	User     string `json:"user"`
}

// SensitiveAccessPayload is the payload for sensitive file access messages.
type SensitiveAccessPayload struct {
	Accesses  []SensitiveAccess `json:"accesses"`
	Total     int               `json:"total"`
	Timestamp int64             `json:"timestamp"`
}

// SensitiveFileCollector checks which processes have sensitive files open.
func SensitiveFileCollector(interval time.Duration) *Collector {
	// Also monitor all users' authorized_keys and .env files
	addUserSensitiveFiles()

	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeSensitiveAccess, SensitiveAccessPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		accesses := scanOpenFiles()
		return TypeSensitiveAccess, SensitiveAccessPayload{
			Accesses:  accesses,
			Total:     len(accesses),
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

func addUserSensitiveFiles() {
	// Add all users' authorized_keys and .env
	entries, err := os.ReadDir("/home")
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		home := filepath.Join("/home", e.Name())
		sensitiveFiles[filepath.Join(home, ".ssh/authorized_keys")] = "user SSH keys"
		sensitiveFiles[filepath.Join(home, ".env")] = "user environment secrets"
	}
}

func scanOpenFiles() []SensitiveAccess {
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var accesses []SensitiveAccess

	for _, entry := range procEntries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		var processName, user string
		nameResolved := false

		for _, fd := range fds {
			target, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}

			reason, ok := sensitiveFiles[target]
			if !ok {
				continue
			}

			// Lazy resolve process name and user
			if !nameResolved {
				processName = readCommName(filepath.Join("/proc", entry.Name()))
				user = readProcUser(filepath.Join("/proc", entry.Name(), "status"))
				nameResolved = true
			}

			accesses = append(accesses, SensitiveAccess{
				PID:     pid,
				Process: processName,
				File:    target,
				Reason:  reason,
				User:    user,
			})
		}
	}

	return accesses
}

func readCommName(procDir string) string {
	data, err := os.ReadFile(filepath.Join(procDir, "comm"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
