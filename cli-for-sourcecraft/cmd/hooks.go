// cmd/hooks.go
package cmd

import "github.com/spf13/cobra"

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage Git hooks integration",              // Simple string
	Long:  "Install Git hooks that call src commands.", // Simple string
}

func init() {
	rootCmd.AddCommand(hooksCmd)
}
