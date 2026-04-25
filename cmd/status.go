package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check project services against portman.yml registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking project status (not yet implemented)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
