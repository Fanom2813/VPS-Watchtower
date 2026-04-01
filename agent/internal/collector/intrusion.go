package collector

import (
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

const TypeIntrusion = "metrics:intrusion"

// IntrusionAlert represents a detected security event.
type IntrusionAlert struct {
	Type      string `json:"type"`      // "brute_force", "priv_escalation", "port_scan"
	Severity  string `json:"severity"`  // "low", "medium", "high", "critical"
	Source    string `json:"source"`    // source IP or process
	Detail    string `json:"detail"`    // human-readable description
	Count     int    `json:"count"`     // number of events (e.g. failed attempts)
	Timestamp int64  `json:"timestamp"`
}

// IntrusionPayload is the payload for intrusion detection messages.
type IntrusionPayload struct {
	Alerts    []IntrusionAlert `json:"alerts"`
	Timestamp int64            `json:"timestamp"`
}

// Thresholds for detection.
const (
	bruteForceThreshold = 5               // failed attempts from same IP
	bruteForceWindow    = 5 * time.Minute // within this time window
)

var (
	reFailedAttempt = regexp.MustCompile(`sshd\[\d+\]: Failed \S+ for (?:invalid user )?(\S+) from (\S+)`)
	reSudo          = regexp.MustCompile(`sudo:\s+(\S+)\s+:.*COMMAND=(.*)`)
	reSuFailed      = regexp.MustCompile(`su\[\d+\]: (?:FAILED SU|pam_authenticate).*`)
)

// IntrusionCollector analyzes auth logs for security events.
func IntrusionCollector(interval time.Duration) *Collector {
	tracker := &intrusionTracker{
		failedAttempts: make(map[string][]time.Time),
	}
	var lastSize int64
	logPath := detectAuthLogPath()

	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" || logPath == "" {
			return TypeIntrusion, IntrusionPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		lines, newSize := readNewLines(logPath, lastSize)
		lastSize = newSize

		alerts := tracker.analyze(lines)

		return TypeIntrusion, IntrusionPayload{
			Alerts:    alerts,
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

type intrusionTracker struct {
	mu             sync.Mutex
	failedAttempts map[string][]time.Time // IP → timestamps of failed attempts
	alerted        map[string]bool        // IPs already alerted for brute force
}

func (t *intrusionTracker) analyze(lines []string) []IntrusionAlert {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.alerted == nil {
		t.alerted = make(map[string]bool)
	}

	now := time.Now()
	var alerts []IntrusionAlert

	for _, line := range lines {
		// Brute-force SSH detection
		if m := reFailedAttempt.FindStringSubmatch(line); m != nil {
			ip := m[2]
			t.failedAttempts[ip] = append(t.failedAttempts[ip], now)
		}

		// Privilege escalation: sudo usage
		if m := reSudo.FindStringSubmatch(line); m != nil {
			user := m[1]
			cmd := strings.TrimSpace(m[2])
			alerts = append(alerts, IntrusionAlert{
				Type:      "priv_escalation",
				Severity:  "medium",
				Source:    user,
				Detail:    "sudo: " + cmd,
				Count:     1,
				Timestamp: now.UnixMilli(),
			})
		}

		// Failed su attempts
		if reSuFailed.MatchString(line) {
			alerts = append(alerts, IntrusionAlert{
				Type:      "priv_escalation",
				Severity:  "high",
				Source:    "su",
				Detail:    "failed su attempt",
				Count:     1,
				Timestamp: now.UnixMilli(),
			})
		}
	}

	// Check for brute force (N failed attempts from same IP in window)
	cutoff := now.Add(-bruteForceWindow)
	for ip, times := range t.failedAttempts {
		// Prune old entries
		var recent []time.Time
		for _, ts := range times {
			if ts.After(cutoff) {
				recent = append(recent, ts)
			}
		}
		t.failedAttempts[ip] = recent

		if len(recent) >= bruteForceThreshold && !t.alerted[ip] {
			t.alerted[ip] = true
			severity := "high"
			if len(recent) >= 20 {
				severity = "critical"
			}
			alerts = append(alerts, IntrusionAlert{
				Type:      "brute_force",
				Severity:  severity,
				Source:    ip,
				Detail:    "SSH brute force detected",
				Count:     len(recent),
				Timestamp: now.UnixMilli(),
			})
		}
	}

	// Reset alerted status when attempts stop
	for ip := range t.alerted {
		if len(t.failedAttempts[ip]) < bruteForceThreshold {
			delete(t.alerted, ip)
		}
	}

	return alerts
}

func readNewLines(path string, lastSize int64) ([]string, int64) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, lastSize
	}

	currentSize := info.Size()
	if currentSize < lastSize || lastSize == 0 {
		return nil, currentSize
	}
	if currentSize == lastSize {
		return nil, lastSize
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, lastSize
	}
	defer f.Close()

	f.Seek(lastSize, 0)
	buf := make([]byte, currentSize-lastSize)
	n, err := f.Read(buf)
	if err != nil {
		return nil, lastSize
	}

	return strings.Split(string(buf[:n]), "\n"), currentSize
}
