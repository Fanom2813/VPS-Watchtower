package main

import (
	"fmt"
	"os"
)

// Version is set at build time via -ldflags.
var Version = "dev"

const usage = `Usage: eyes-agent <command> [options]

Commands:
  setup      Configure the agent and generate a pairing token
  token      Generate a new pairing token
  uninstall  Remove the systemd service (requires root)
  status     Show service status
  run        Start the agent server (used by systemd)

Quick start:
  1. eyes-agent setup --port 9090
  2. Copy the pairing token and agent URL into the desktop app
  3. eyes-agent run

Production (persists across reboots):
  sudo eyes-agent setup --port 9090 --persist

Examples:
  eyes-agent setup --port 9090
  sudo eyes-agent setup --port 9090 --persist
  eyes-agent token
  eyes-agent status
  sudo eyes-agent uninstall
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup(os.Args[2:])
	case "token":
		cmdToken(os.Args[2:])
	case "uninstall":
		cmdUninstall()
	case "status":
		cmdStatus()
	case "run":
		cmdRun(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(1)
	}
}
