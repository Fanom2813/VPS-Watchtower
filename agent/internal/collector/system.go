package collector

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const TypeSystemMetrics = "metrics:system"

// SystemMetrics contains current system resource usage.
type SystemMetrics struct {
	CPUPercent    float64       `json:"cpuPercent"`
	MemTotal      uint64        `json:"memTotal"`
	MemUsed       uint64        `json:"memUsed"`
	MemPercent    float64       `json:"memPercent"`
	DiskTotal     uint64        `json:"diskTotal"`
	DiskUsed      uint64        `json:"diskUsed"`
	DiskPercent   float64       `json:"diskPercent"`
	Uptime        uint64        `json:"uptime"`
	LoadAvg       [3]float64    `json:"loadAvg"`
	NetRxBytes    uint64        `json:"netRxBytes"`
	NetTxBytes    uint64        `json:"netTxBytes"`
	Timestamp     int64         `json:"timestamp"`
}

// SystemCollector creates a collector that gathers system metrics.
func SystemCollector(interval time.Duration) *Collector {
	var prevIdle, prevTotal uint64
	var prevNetRx, prevNetTx uint64
	var firstRun = true

	return New(interval, func() (string, any, error) {
		m := SystemMetrics{Timestamp: time.Now().UnixMilli()}

		if runtime.GOOS == "linux" {
			// CPU
			idle, total := readCPU()
			if !firstRun && total > prevTotal {
				deltaTotal := total - prevTotal
				deltaIdle := idle - prevIdle
				m.CPUPercent = float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
			}
			prevIdle, prevTotal = idle, total

			// Memory
			m.MemTotal, m.MemUsed = readMemory()
			if m.MemTotal > 0 {
				m.MemPercent = float64(m.MemUsed) / float64(m.MemTotal) * 100
			}

			// Disk
			m.DiskTotal, m.DiskUsed = readDisk()
			if m.DiskTotal > 0 {
				m.DiskPercent = float64(m.DiskUsed) / float64(m.DiskTotal) * 100
			}

			// Uptime
			m.Uptime = readUptime()

			// Load average
			m.LoadAvg = readLoadAvg()

			// Network
			rx, tx := readNetworkTotal()
			if !firstRun {
				m.NetRxBytes = rx - prevNetRx
				m.NetTxBytes = tx - prevNetTx
			}
			prevNetRx, prevNetTx = rx, tx
		}

		firstRun = false
		return TypeSystemMetrics, m, nil
	})
}

// --- Linux /proc parsers ---

func readCPU() (idle, total uint64) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0
	}
	for i, f := range fields[1:] {
		v, _ := strconv.ParseUint(f, 10, 64)
		total += v
		if i == 3 { // idle is the 4th field
			idle = v
		}
	}
	return idle, total
}

func readMemory() (total, used uint64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	var memTotal, memAvailable uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(fields[1], 10, 64)
		v *= 1024 // kB to bytes
		switch fields[0] {
		case "MemTotal:":
			memTotal = v
		case "MemAvailable:":
			memAvailable = v
		}
	}
	return memTotal, memTotal - memAvailable
}

func readDisk() (total, used uint64) {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return 0, 0
	}
	// Find the root mount
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "/" {
			return statFS(fields[1])
		}
	}
	return statFS("/")
}

func readUptime() uint64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0
	}
	v, _ := strconv.ParseFloat(fields[0], 64)
	return uint64(v)
}

func readLoadAvg() [3]float64 {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return [3]float64{}
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return [3]float64{}
	}
	var avg [3]float64
	for i := 0; i < 3; i++ {
		avg[i], _ = strconv.ParseFloat(fields[i], 64)
	}
	return avg
}

func readNetworkTotal() (rx, tx uint64) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, ":") || strings.HasPrefix(line, "lo:") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}
		r, _ := strconv.ParseUint(fields[0], 10, 64)
		t, _ := strconv.ParseUint(fields[8], 10, 64)
		rx += r
		tx += t
	}
	return rx, tx
}
