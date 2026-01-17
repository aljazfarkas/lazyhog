package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/config"
	"github.com/aljazfarkas/lazyhog/internal/ui/miller"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var debugFlag bool

var rootCmd = &cobra.Command{
	Use:   "lazyhog",
	Short: "PostHog TUI - Terminal interface for PostHog",
	Long: `lazyhog is a blazing-fast Terminal User Interface for PostHog.

Features:
  • Miller Columns interface for rapid debugging workflows
  • Live event streaming with smart polling
  • Instant event → person pivot for debugging
  • Feature flag management
  • Responsive design for any terminal width

Get started by running: lazyhog login

Run without arguments to launch the Miller Columns interface.`,
	Version: version,
	RunE:    runMillerColumns,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetVersionTemplate(fmt.Sprintf("lazyhog version %s\n", version))
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging (shows full request/response details)")
}

func runMillerColumns(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'lazyhog login' to set up authentication", err)
	}

	// Set debug mode from flag
	cfg.Debug = debugFlag

	// Create client
	c := client.New(cfg)

	// Initialize project context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = c.InitializeProject(ctx)
	cancel()
	if err != nil {
		// Don't fail - allow TUI to start and use fallback
		fmt.Fprintf(os.Stderr, "Warning: Could not initialize project: %v\n", err)
		fmt.Fprintf(os.Stderr, "Starting TUI anyway...\n")
		time.Sleep(1 * time.Second)
	}

	// Show debug logging info if enabled
	if debugFlag {
		logPath := filepath.Join(os.Getenv("HOME"), ".config", "lazyhog-debug.log")
		fmt.Fprintf(os.Stderr, "Debug logging enabled: %s\n", logPath)
		fmt.Fprintf(os.Stderr, "Tip: In another terminal run: tail -f %s\n\n", logPath)
		time.Sleep(2 * time.Second) // Give user time to read
	}

	// Ensure client resources are cleaned up
	defer c.Close()

	// Create and run the Miller Columns TUI
	m := miller.New(c)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running lazyhog: %w", err)
	}

	return nil
}
