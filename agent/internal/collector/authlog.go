package collector

import (
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const TypeAuthLog = "metrics:authlog"

// AuthEntry represents a single auth log event.
type AuthEntry struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"` // "login_success", "login_failed", "disconnect", "invalid_user"
	User      string `json:"user"`
	Source    string `json:"source"` // IP address
	Method    string `json:"method"` // "password", "publickey", etc.
	Raw       string `json:"raw"`
}

// AuthLogPayload is the payload for auth log messages.
type AuthLogPayload struct {
	Entries   []AuthEntry `json:"entries"`
	Timestamp int64       `json:"timestamp"`
}

var (
	// sshd[12345]: Accepted publickey for root from 1.2.3.4 port 22 ssh2
	reAccepted = regexp.MustCompile(`sshd\[\d+\]: Accepted (\S+) for (\S+) from (\S+)`)
	// sshd[12345]: Failed password for root from 1.2.3.4 port 22 ssh2
	reFailed = regexp.MustCompile(`sshd\[\d+\]: Failed (\S+) for (\S+) from (\S+)`)
	// sshd[12345]: Failed password for invalid user admin from 1.2.3.4
	reInvalidUser = regexp.MustCompile(`sshd\[\d+\]: Failed (\S+) for invalid user (\S+) from (\S+)`)
	// sshd[12345]: Invalid user admin from 1.2.3.4
	reInvalid = regexp.MustCompile(`sshd\[\d+\]: Invalid user (\S+) from (\S+)`)
	// sshd[12345]: Disconnected from user root 1.2.3.4 port 22
	reDisconnect = regexp.MustCompile(`sshd\[\d+\]: Disconnected from user (\S+) (\S+)`)
)

// AuthLogCollector tails the auth log and sends new entries.
func AuthLogCollector(interval time.Duration) *Collector {
	var lastSize int64

	logPath := detectAuthLogPath()

	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" || logPath == "" {
			return TypeAuthLog, AuthLogPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		entries, newSize := readNewAuthEntries(logPath, lastSize)
		lastSize = newSize

		return TypeAuthLog, AuthLogPayload{
			Entries:   entries,
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

func detectAuthLogPath() string {
	paths := []string{
		"/var/log/auth.log",   // Debian/Ubuntu
		"/var/log/secure",     // RHEL/CentOS
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func readNewAuthEntries(path string, lastSize int64) ([]AuthEntry, int64) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, lastSize
	}

	currentSize := info.Size()

	// File rotated or first run — start from current position
	if currentSize < lastSize || lastSize == 0 {
		return nil, currentSize
	}

	// Nothing new
	if currentSize == lastSize {
		return nil, lastSize
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, lastSize
	}
	defer f.Close()

	// Seek to where we left off
	f.Seek(lastSize, 0)

	buf := make([]byte, currentSize-lastSize)
	n, err := f.Read(buf)
	if err != nil {
		return nil, lastSize
	}

	lines := strings.Split(string(buf[:n]), "\n")
	var entries []AuthEntry

	for _, line := range lines {
		if entry, ok := parseAuthLine(line); ok {
			entries = append(entries, entry)
		}
	}

	return entries, currentSize
}

func parseAuthLine(line string) (AuthEntry, bool) {
	if !strings.Contains(line, "sshd[") {
		return AuthEntry{}, false
	}

	// Extract timestamp (first 3 fields: "Mon DD HH:MM:SS")
	ts := extractTimestamp(line)

	if m := reInvalidUser.FindStringSubmatch(line); m != nil {
		return AuthEntry{
			Timestamp: ts,
			Type:      "login_failed",
			Method:    m[1],
			User:      m[2],
			Source:    m[3],
			Raw:       line,
		}, true
	}

	if m := reAccepted.FindStringSubmatch(line); m != nil {
		return AuthEntry{
			Timestamp: ts,
			Type:      "login_success",
			Method:    m[1],
			User:      m[2],
			Source:    m[3],
			Raw:       line,
		}, true
	}

	if m := reFailed.FindStringSubmatch(line); m != nil {
		return AuthEntry{
			Timestamp: ts,
			Type:      "login_failed",
			Method:    m[1],
			User:      m[2],
			Source:    m[3],
			Raw:       line,
		}, true
	}

	if m := reInvalid.FindStringSubmatch(line); m != nil {
		return AuthEntry{
			Timestamp: ts,
			Type:      "invalid_user",
			User:      m[1],
			Source:    m[2],
			Raw:       line,
		}, true
	}

	if m := reDisconnect.FindStringSubmatch(line); m != nil {
		return AuthEntry{
			Timestamp: ts,
			Type:      "disconnect",
			User:      m[1],
			Source:    m[2],
			Raw:       line,
		}, true
	}

	return AuthEntry{}, false
}

func extractTimestamp(line string) string {
	fields := strings.Fields(line)
	if len(fields) >= 3 {
		return fields[0] + " " + fields[1] + " " + fields[2]
	}
	return ""
}
