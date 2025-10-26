// cmd/repo.go
package cmd

import "github.com/spf13/cobra"

// repoCmd - родительская команда 'src repo'
var repoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Repos ",
	Aliases: []string{"repository"},
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
