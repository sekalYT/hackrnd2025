// cmd/access_role_list.go
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	// Need api types
	"cli-for-sourcecraft/internal/git" // To potentially detect repo

	"github.com/spf13/cobra"
)

var roleListRepoFlag string // Specific flag for this command

var roleListCmd = &cobra.Command{
	Use:   "list [flags]", // No <repo> arg, use flag or detect
	Short: "List user roles in a repository",
	Long:  `Displays a list of users and their assigned roles for a specific repository.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var orgSlug, repoSlug string
		var err error

		// 1. Determine Repository
		if roleListRepoFlag != "" {
			parts := strings.SplitN(roleListRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid format for --repo flag: '%s'. Expected: <org>/<repo>", roleListRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			// Try to detect from git remote
			fmt.Println("Attempting to detect repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("could not detect repository. Use the --repo <org>/<repo> flag")
			}
		}
		fmt.Printf("Fetching roles for repository: %s/%s\n", orgSlug, repoSlug)

		// 2. Call API
		roles, err := apiClient.ListRepoRoles(orgSlug, repoSlug)
		if err != nil {
			return err // API error (404, 403, etc.)
		}

		if len(roles) == 0 {
			fmt.Println("No specific user roles found for this repository.")
			return nil
		}

		// 3. Display Results
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "USER ID\tTYPE\tROLE")
		fmt.Fprintln(w, "-------\t----\t----")

		for _, sr := range roles {
			// Assuming Subject is always present based on Swagger for Add/Remove
			userID := sr.Subject.ID
			userType := sr.Subject.Type
			roleName := sr.Role

			fmt.Fprintf(w, "%s\t%s\t%s\n", userID, userType, roleName)
		}
		return w.Flush()
	},
}

func init() {
	roleCmd.AddCommand(roleListCmd)
	// Add required --repo flag (or make detection mandatory)
	roleListCmd.Flags().StringVarP(&roleListRepoFlag, "repo", "R", "", "Specify repository in <org>/<repo> format (required for this command)")
	roleListCmd.MarkFlagRequired("repo") // Make --repo mandatory as detection might not be suitable for permissions
}
