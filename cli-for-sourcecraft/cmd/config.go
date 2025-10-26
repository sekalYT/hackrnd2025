// cmd/config.go
package cmd

import "github.com/spf13/cobra"

// configCmd - базовая команда 'src config'
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "CLI configuration management",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
