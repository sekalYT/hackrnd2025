// cmd/repo.go
package cmd

import "github.com/spf13/cobra"

// repoCmd - родительская команда 'src repo'
var repoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Repos ",
	Aliases: []string{"repository"}, // Алиас для удобства
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
