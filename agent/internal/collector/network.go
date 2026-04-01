package collector

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const TypeNetworkConnections = "metrics:network"

// NetConnection represents a single TCP connection.
type NetConnection struct {
	LocalAddr  string `json:"localAddr"`
	LocalPort  int    `json:"localPort"`
	RemoteAddr string `json:"remoteAddr"`
	RemotePort int    `json:"remotePort"`
	State      string `json:"state"`
	PID        int    `json:"pid,omitempty"`
}

// NetworkPayload is the payload for network connection messages.
type NetworkPayload struct {
	Connections []NetConnection `json:"connections"`
	Listening   []NetConnection `json:"listening"`
	Total       int             `json:"total"`
	Timestamp   int64           `json:"timestamp"`
}

var tcpStates = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}

// NetworkCollector creates a collector that gathers active TCP connections.
func NetworkCollector(interval time.Duration) *Collector {
	return New(interval, func() (string, any, error) {
		if runtime.GOOS != "linux" {
			return TypeNetworkConnections, NetworkPayload{Timestamp: time.Now().UnixMilli()}, nil
		}

		var allConns []NetConnection
		for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
			conns := readProcNet(path)
			allConns = append(allConns, conns...)
		}

		var listening, active []NetConnection
		for _, c := range allConns {
			if c.State == "LISTEN" {
				listening = append(listening, c)
			} else if c.State == "ESTABLISHED" {
				active = append(active, c)
			}
		}

		return TypeNetworkConnections, NetworkPayload{
			Connections: active,
			Listening:   listening,
			Total:       len(allConns),
			Timestamp:   time.Now().UnixMilli(),
		}, nil
	})
}

func readProcNet(path string) []NetConnection {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return nil
	}

	isV6 := strings.HasSuffix(path, "tcp6")
	var conns []NetConnection

	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		localAddr, localPort := parseAddr(fields[1], isV6)
		remoteAddr, remotePort := parseAddr(fields[2], isV6)
		state := tcpStates[strings.ToUpper(fields[3])]
		if state == "" {
			state = fields[3]
		}

		conns = append(conns, NetConnection{
			LocalAddr:  localAddr,
			LocalPort:  localPort,
			RemoteAddr: remoteAddr,
			RemotePort: remotePort,
			State:      state,
		})
	}

	return conns
}

func parseAddr(s string, isV6 bool) (string, int) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", 0
	}

	port, _ := strconv.ParseInt(parts[1], 16, 32)

	if isV6 {
		return parseIPv6(parts[0]), int(port)
	}
	return parseIPv4(parts[0]), int(port)
}

func parseIPv4(hexStr string) string {
	if len(hexStr) != 8 {
		return hexStr
	}
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != 4 {
		return hexStr
	}
	// /proc/net/tcp uses little-endian
	return fmt.Sprintf("%d.%d.%d.%d", b[3], b[2], b[1], b[0])
}

func parseIPv6(hexStr string) string {
	if len(hexStr) != 32 {
		return hexStr
	}
	b, err := hex.DecodeString(hexStr)
	if err != nil || len(b) != 16 {
		return hexStr
	}

	// Check for IPv4-mapped (::ffff:x.x.x.x)
	allZero := true
	for i := 0; i < 10; i++ {
		if b[i] != 0 {
			allZero = false
			break
		}
	}
	if allZero && b[10] == 0xff && b[11] == 0xff {
		return fmt.Sprintf("%d.%d.%d.%d", b[12], b[13], b[14], b[15])
	}

	// /proc/net/tcp6 uses groups of 4 bytes in little-endian
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		b[3], b[2], b[1], b[0],
		b[7], b[6], b[5], b[4],
		b[11], b[10], b[9], b[8],
		b[15], b[14], b[13], b[12],
	)
}
