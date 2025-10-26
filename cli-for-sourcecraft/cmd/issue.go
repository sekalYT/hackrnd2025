// cmd/issue.go
package cmd

import "github.com/spf13/cobra"

// issueCmd - базовая команда 'src issue'
var issueCmd = &cobra.Command{
	Use:     "issue",
	Short:   "Working with SourceCraft Issues",
	Aliases: []string{"issues"},
}

func init() {
	rootCmd.AddCommand(issueCmd) // Добавляем 'issue' к 'src'
}
