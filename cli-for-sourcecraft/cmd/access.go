// cmd/access.go
package cmd

import "github.com/spf13/cobra"

// accessCmd represents the base command for access control.
var accessCmd = &cobra.Command{
	Use:     "access",
	Short:   "Manage access permissions (e.g., repository roles)",
	Aliases: []string{"permissions"},
}

func init() {
	rootCmd.AddCommand(accessCmd)
}
