package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/eyes-on-vps/agent/internal/auth"
	"github.com/eyes-on-vps/agent/internal/collector"
	"github.com/eyes-on-vps/agent/internal/config"
	"github.com/eyes-on-vps/agent/internal/protocol"
	"github.com/eyes-on-vps/agent/internal/service"
	"github.com/eyes-on-vps/agent/internal/sysinfo"
	"github.com/eyes-on-vps/agent/internal/transport"
)

func cmdSetup(args []string) {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	port := fs.Int("port", 9090, "Port for the WebSocket server")
	configPath := fs.String("config", "agent.json", "Path to config file")
	persist := fs.Bool("persist", false, "Install as a systemd service (requires root)")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	cfg.Port = *port
	cfg.SetPath(*configPath)

	// Generate a new pairing token
	if err := cfg.GeneratePairingToken(); err != nil {
		log.Fatalf("generate pairing token: %v", err)
	}

	if err := cfg.Save(); err != nil {
		log.Fatalf("save config: %v", err)
	}

	fmt.Println("Agent configured successfully!")
	fmt.Printf("  Config:   %s\n", *configPath)
	fmt.Printf("  Agent ID: %s\n", cfg.AgentID)
	fmt.Printf("  Port:     %d\n", cfg.Port)
	fmt.Println()
	ip := sysinfo.DetectIP()
	fmt.Println()
	fmt.Printf("Agent URL:      ws://%s:%d/ws\n", ip, cfg.Port)
	fmt.Printf("Pairing Token:  %s\n", cfg.PairingToken)

	if !*persist {
		fmt.Println()
		fmt.Println("Run 'eyes-agent run' to start in foreground.")
		fmt.Println("Or re-run with --persist to install as a systemd service.")
		return
	}

	binary, err := os.Executable()
	if err != nil {
		log.Fatalf("resolve binary path: %v", err)
	}

	fmt.Println()
	fmt.Println("Installing systemd service...")

	if err := service.Install(binary); err != nil {
		log.Fatalf("install failed: %v", err)
	}

	if err := service.Start(); err != nil {
		log.Fatalf("start failed: %v", err)
	}

	fmt.Println()
	fmt.Println("eyes-agent installed and running!")
	fmt.Printf("  Binary:  %s\n", service.BinaryPath())
	fmt.Printf("  Config:  %s\n", service.ConfigPath())
	fmt.Printf("  Service: %s\n", service.ServiceName)
	fmt.Println()
	fmt.Println("Useful commands:")
	fmt.Println("  eyes-agent status                    Show service status")
	fmt.Println("  sudo systemctl restart eyes-agent    Restart the service")
	fmt.Println("  sudo journalctl -u eyes-agent -f     Follow logs")
	fmt.Println("  sudo eyes-agent uninstall            Remove the service")
}

func cmdToken(args []string) {
	fs := flag.NewFlagSet("token", flag.ExitOnError)
	configPath := fs.String("config", "agent.json", "Path to config file")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := cfg.GeneratePairingToken(); err != nil {
		log.Fatalf("generate token: %v", err)
	}

	if err := cfg.Save(); err != nil {
		log.Fatalf("save config: %v", err)
	}

	fmt.Println("New pairing token generated:")
	fmt.Printf("  %s\n", cfg.PairingToken)
}

func cmdUninstall() {
	fmt.Println("Uninstalling eyes-agent service...")

	if err := service.Uninstall(); err != nil {
		log.Fatalf("uninstall failed: %v", err)
	}

	fmt.Println("Service removed.")
	fmt.Printf("Config preserved at %s — delete manually if no longer needed.\n", service.ConfigPath())
}

func cmdStatus() {
	output, _ := service.Status()
	fmt.Print(output)
}

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", "agent.json", "Path to config file")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.Port == 0 {
		log.Fatal("port not set — run 'eyes-agent setup' first")
	}

	authHandler := auth.NewHandler(cfg, Version)
	server := transport.NewServer(authHandler, func(msg protocol.Message) {
		log.Printf("received: type=%s", msg.Type)
	})

	// Collectors only run while at least one desktop is connected
	metrics := collector.NewManager(
		collector.SystemCollector(3*time.Second),
		collector.ProcessCollector(10*time.Second),
		collector.AuthLogCollector(5*time.Second),
		collector.NetworkCollector(10*time.Second),
		collector.DockerCollector(15*time.Second),
		collector.IntrusionCollector(5*time.Second),
		collector.OutboundCollector(10*time.Second),
		collector.TamperCollector(60*time.Second),
		collector.CronCollector(30*time.Second),
		collector.SensitiveFileCollector(15*time.Second),
		collector.SystemdCollector(30*time.Second),
		collector.FileIntegrityCollector(30*time.Second),
	)
	server.OnActive(metrics.HandleActive)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("received %s — shutting down", sig)
		cancel()
	}()

	log.Printf("agent %s starting on port %d", cfg.AgentID, cfg.Port)

	if err := server.Run(ctx, cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
