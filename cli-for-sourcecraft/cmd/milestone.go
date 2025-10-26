// cmd/milestone.go
package cmd

import "github.com/spf13/cobra"

// milestoneCmd - базовая команда 'src milestone'
var milestoneCmd = &cobra.Command{
	Use:     "milestone",
	Short:   "Working with Milestones (SourceCraft)",
	Aliases: []string{"milestones"},
}

func init() {
	rootCmd.AddCommand(milestoneCmd) // Добавляем 'milestone' к 'src'
}
