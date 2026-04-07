package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/monitor"
	"github.com/user/portwatch/internal/notifier"
)

const version = "0.1.0"

func main() {
	// CLI flags
	configPath := flag.String("config", "portwatch.yaml", "path to config file")
	interval := flag.Int("interval", 30, "scan interval in seconds")
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("portwatch v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		// Config file is optional; fall back to defaults with CLI overrides
		if *verbose {
			fmt.Fprintf(os.Stderr, "[warn] could not load config file %q: %v — using defaults\n", *configPath, err)
		}
		cfg = config.Default()
	}

	// CLI flags override config file values when explicitly set
	if *interval != 30 {
		cfg.IntervalSeconds = *interval
	}
	cfg.Verbose = *verbose

	// Build notifier chain based on config
	n, err := notifier.Build(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] failed to initialise notifiers: %v\n", err)
		os.Exit(1)
	}

	// Create and start the port monitor
	m := monitor.New(cfg, n)

	if cfg.Verbose {
		fmt.Printf("portwatch v%s starting (interval: %ds)\n", version, cfg.IntervalSeconds)
	}

	if err := m.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] monitor failed to start: %v\n", err)
		os.Exit(1)
	}

	// Block until SIGINT or SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nshutting down portwatch...")
	m.Stop()
}
