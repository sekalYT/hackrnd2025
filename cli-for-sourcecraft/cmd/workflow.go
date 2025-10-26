// cmd/workflow.go
package cmd

import "github.com/spf13/cobra"

// workflowCmd - базовая команда 'src workflow'
var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Short:   "Working with SourceCraft CI/CD Workflows",
	Aliases: []string{"workflows", "wf"},
	Long:    `Launch CI/CD Workflows`,
}

func init() {
	rootCmd.AddCommand(workflowCmd) // Добавляем 'workflow' к 'src'
}
