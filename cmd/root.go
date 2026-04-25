package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "portman",
	Short: "A clean CLI for managing local ports and processes",
	Long: `portman lets you list listening ports, check availability,
kill processes by port, find next free ports, and monitor project services.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
