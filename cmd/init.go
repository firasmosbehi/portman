package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initForceFlag bool
var initBlankFlag bool

const defaultPortmanYML = `services:
  - name: web
    port: 3000
    command: npm run dev

  - name: api
    port: 3001
    command: npm run api

  - name: db
    port: 5432
    health_check: pg_isready
`

const blankPortmanYML = `services: []
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a sample portman.yml in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		const filename = "portman.yml"

		if _, err := os.Stat(filename); err == nil && !initForceFlag {
			return fmt.Errorf("portman.yml already exists. Use --force to overwrite")
		}

		content := defaultPortmanYML
		if initBlankFlag {
			content = blankPortmanYML
		}

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write portman.yml: %w", err)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", filename)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initForceFlag, "force", "f", false, "Overwrite existing portman.yml")
	initCmd.Flags().BoolVarP(&initBlankFlag, "blank", "b", false, "Generate an empty template")
	rootCmd.AddCommand(initCmd)
}
