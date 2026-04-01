package sysinfo

import (
	"net"
	"os"
	"runtime"
	"strings"
)

// StaticInfo contains machine identity fields that don't change.
type StaticInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Distro   string `json:"distro"`
	Hostname string `json:"hostname"`
}

// Collect gathers static system information.
func Collect() StaticInfo {
	hostname, _ := os.Hostname()

	return StaticInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Distro:   detectDistro(),
		Hostname: hostname,
	}
}

// DetectIP returns the first non-loopback IPv4 address.
func DetectIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				return ip.String()
			}
		}
	}
	return "127.0.0.1"
}

func detectDistro() string {
	if runtime.GOOS != "linux" {
		return runtime.GOOS
	}

	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "linux"
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			val := strings.TrimPrefix(line, "PRETTY_NAME=")
			val = strings.Trim(val, "\"")
			return val
		}
	}

	return "linux"
}
