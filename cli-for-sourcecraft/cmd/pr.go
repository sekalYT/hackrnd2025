// cmd/pr.go
package cmd

import "github.com/spf13/cobra"

// prCmd - базовая команда 'src pr'
var prCmd = &cobra.Command{
	Use:     "pr",
	Short:   "Working with SourceCraft Pull Requests",
	Aliases: []string{"pullrequest", "pull"},
}

func init() {
	rootCmd.AddCommand(prCmd) // Добавляем 'pr' к 'src'
}
