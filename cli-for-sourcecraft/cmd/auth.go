// cmd/auth.go
package cmd

import "github.com/spf13/cobra"

// authCmd - базовая команда 'src auth'
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
}

func init() {
	rootCmd.AddCommand(authCmd) // Добавляем 'auth' к 'src'
}
