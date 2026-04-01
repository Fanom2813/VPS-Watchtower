package collector

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const TypeProcessList = "metrics:processes"

// Process classification types.
const (
	ClassKernel  = "kernel"  // kernel threads (ppid 0 or 2)
	ClassSystem  = "system"  // binary in standard system paths
	ClassUnknown = "unknown" // everything else — potential concern
)

// Standard system binary directories.
var systemPaths = []string{
	"/usr/bin/", "/usr/sbin/", "/bin/", "/sbin/",
	"/usr/lib/", "/usr/libexec/", "/lib/",
	"/usr/local/bin/", "/usr/local/sbin/",
	"/snap/", "/opt/",
}

// ProcessInfo represents a single running process.
type ProcessInfo struct {
	PID      int    `json:"pid"`
	PPID     int    `json:"ppid"`
	Name     string `json:"name"`
	BinPath  string `json:"binPath"`
	Cmdline  string `json:"cmdline"`
	State    string `json:"state"`
	User     string `json:"user"`
	Class    string `json:"class"` // "kernel", "system", "unknown"
	CPUTime  uint64 `json:"cpuTime"`
	MemRSS   uint64 `json:"memRss"`
}

// ProcessList is the payload for the process list message.
type ProcessList struct {
	Processes []ProcessInfo `json:"processes"`
	Total     int           `json:"total"`
	Unknown   int           `json:"unknown"` // count of unclassified processes
	Timestamp int64         `json:"timestamp"`
}

// ProcessCollector creates a collector that gathers the process list.
func ProcessCollector(interval time.Duration) *Collector {
	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeProcessList, ProcessList{Timestamp: time.Now().UnixMilli()}, nil
		}

		procs := readProcesses()
		unknown := 0
		for _, p := range procs {
			if p.Class == ClassUnknown {
				unknown++
			}
		}

		return TypeProcessList, ProcessList{
			Processes: procs,
			Total:     len(procs),
			Unknown:   unknown,
			Timestamp: time.Now().UnixMilli(),
		}, nil
	})
}

func readProcesses() []ProcessInfo {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var procs []ProcessInfo
	pageSize := uint64(os.Getpagesize())

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		procDir := filepath.Join("/proc", entry.Name())

		statPath := filepath.Join(procDir, "stat")
		data, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		p := parseProcStat(string(data), pid, pageSize)
		if p.Name == "" {
			continue
		}

		// Binary path from /proc/pid/exe symlink
		p.BinPath = readExePath(procDir)

		// Command line
		p.Cmdline = readCmdline(procDir)

		// User (UID)
		p.User = readProcUser(filepath.Join(procDir, "status"))

		// Classify
		p.Class = classifyProcess(p)

		procs = append(procs, p)
	}

	return procs
}

func parseProcStat(data string, pid int, pageSize uint64) ProcessInfo {
	start := strings.IndexByte(data, '(')
	end := strings.LastIndexByte(data, ')')
	if start < 0 || end < 0 || end <= start {
		return ProcessInfo{}
	}

	name := data[start+1 : end]
	rest := strings.Fields(data[end+2:])

	if len(rest) < 22 {
		return ProcessInfo{}
	}

	state := rest[0]
	ppid, _ := strconv.Atoi(rest[1])
	utime, _ := strconv.ParseUint(rest[11], 10, 64)
	stime, _ := strconv.ParseUint(rest[12], 10, 64)
	rss, _ := strconv.ParseUint(rest[21], 10, 64)

	return ProcessInfo{
		PID:     pid,
		PPID:    ppid,
		Name:    name,
		State:   state,
		CPUTime: utime + stime,
		MemRSS:  rss * pageSize,
	}
}

func readExePath(procDir string) string {
	target, err := os.Readlink(filepath.Join(procDir, "exe"))
	if err != nil {
		return ""
	}
	// Kernel may append " (deleted)" for removed binaries
	target = strings.TrimSuffix(target, " (deleted)")
	return target
}

func readCmdline(procDir string) string {
	data, err := os.ReadFile(filepath.Join(procDir, "cmdline"))
	if err != nil || len(data) == 0 {
		return ""
	}
	// cmdline uses null bytes as separators
	return strings.Join(strings.Split(strings.TrimRight(string(data), "\x00"), "\x00"), " ")
}

func readProcUser(statusPath string) string {
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

func classifyProcess(p ProcessInfo) string {
	// Kernel threads: PID 1's children with no exe, or PPID 0/2
	if p.PPID == 0 || p.PPID == 2 || p.PID == 2 {
		return ClassKernel
	}
	if p.BinPath == "" {
		// No exe link — likely a kernel thread
		return ClassKernel
	}

	for _, prefix := range systemPaths {
		if strings.HasPrefix(p.BinPath, prefix) {
			return ClassSystem
		}
	}

	return ClassUnknown
}
