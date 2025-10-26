// cmd/access_role.go
package cmd

import "github.com/spf13/cobra"

// roleCmd represents the command for managing roles.
var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage repository roles for users",
	Long:  `List, add, or remove user roles within a specific repository.`,
}

func init() {
	accessCmd.AddCommand(roleCmd) // Add 'role' to 'access'
}
