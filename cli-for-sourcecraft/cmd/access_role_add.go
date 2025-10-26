// cmd/access_role_add.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/api" // Need api.RepoRole type

	"github.com/spf13/cobra"
)

// Define allowed roles based on Swagger/API constants
var allowedRoles = []string{
	string(api.RepoRoleViewer),
	string(api.RepoRoleContributor),
	string(api.RepoRoleDeveloper),
	string(api.RepoRoleMaintainer),
	string(api.RepoRoleAdmin),
}

var roleAddCmd = &cobra.Command{
	Use:   "add <repository> <user_id> <role>",
	Short: "Add a role for a user in a repository",
	Long: fmt.Sprintf(`Grants a specified role to a user within a repository.

<repository>: Repository path in <org>/<repo> format.
<user_id>: The UUID of the user.
<role>: The role to assign. Allowed values: %s`, strings.Join(allowedRoles, ", ")),
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
		fmt.Printf("Adding role '%s' for user '%s' in repository '%s/%s'...\n", role, userID, orgSlug, repoSlug)
		err := apiClient.AddRepoRole(orgSlug, repoSlug, userID, role)
		if err != nil {
			return fmt.Errorf("failed to add role: %w", err) // Wrap API error
		}

		fmt.Println("Role added successfully.")
		return nil
	},
}

func init() {
	roleCmd.AddCommand(roleAddCmd)
	// No extra flags needed for this command
}
