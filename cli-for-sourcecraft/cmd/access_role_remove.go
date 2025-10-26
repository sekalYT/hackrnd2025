// cmd/access_role_remove.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/api" // Need api.RepoRole type

	"github.com/spf13/cobra"
)

// allowedRoles is already defined in access_role_add.go within the same package

var roleRemoveCmd = &cobra.Command{
	Use:   "remove <repository> <user_id> <role>",
	Short: "Remove a role from a user in a repository",
	Long: fmt.Sprintf(`Removes a specified role from a user within a repository.

<repository>: Repository path in <org>/<repo> format.
<user_id>: The UUID of the user.
<role>: The role to remove. Allowed values: %s`, strings.Join(allowedRoles, ", ")),
	Args: cobra.ExactArgs(3), // Requires repo, user_id, role
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := args[0]
		userID := args[1]
		roleStr := args[2]

		// 1. Parse Repository Path
		parts := strings.SplitN(repoPath, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid repository format: '%s'. Expected: <org>/<repo>", repoPath)
		}
		orgSlug := parts[0]
		repoSlug := parts[1]

		// 2. Validate Role
		isValidRole := false
		for _, r := range allowedRoles {
			if roleStr == r {
				isValidRole = true
				break
			}
		}
		if !isValidRole {
			return fmt.Errorf("invalid role: '%s'. Allowed roles are: %s", roleStr, strings.Join(allowedRoles, ", "))
		}
		role := api.RepoRole(roleStr) // Convert validated string to type

		// 3. Call API
		fmt.Printf("Removing role '%s' for user '%s' in repository '%s/%s'...\n", role, userID, orgSlug, repoSlug)
		err := apiClient.RemoveRepoRole(orgSlug, repoSlug, userID, role)
		if err != nil {
			return fmt.Errorf("failed to remove role: %w", err) // Wrap API error
		}

		fmt.Println("Role removed successfully.")
		return nil
	},
}

func init() {
	roleCmd.AddCommand(roleRemoveCmd)
	// No extra flags needed for this command
}
