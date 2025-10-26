// cmd/issue_list.go
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	issueListRepoFlag string
)

var issueListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "View the list of tasks in the repository",
	Long:  `Shows the list of issues for the specified repository.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		var orgSlug, repoSlug string
		var err error

		if issueListRepoFlag != "" {
			parts := strings.SplitN(issueListRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", issueListRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
			fmt.Printf("Search for issues in the specified repository: %s/%s\n", orgSlug, repoSlug)
		} else {
			fmt.Println("Defining a repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("could not identify the repository from git remote and 'organization' is not set in the config. Use the --repo <org>/<repo>")
				}
				return fmt.Errorf("failed to identify the repository slug from git remote. Use the --repo <org>/<repo>")
			}
			fmt.Printf("Repository defined: %s/%s\n", orgSlug, repoSlug)
		}

		fmt.Printf("Request issues for %s/%s...\n", orgSlug, repoSlug)
		issues, err := apiClient.ListRepositoryIssues(orgSlug, repoSlug)
		if err != nil {
			return err
		}

		if len(issues) == 0 {
			fmt.Println("Issues not found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tASSIGNEE\tPRIORITY\tUPDATED")
		fmt.Fprintln(w, "--\t-----\t------\t--------\t--------\t-------")

		for _, issue := range issues {
			issueID := cliutils.DerefString(issue.Slug)
			title := cliutils.DerefString(issue.Title)
			status := ""
			if issue.Status != nil {
				status = cliutils.DerefString(issue.Status.Name)
			}
			assignee := ""
			if issue.Assignee != nil {
				assignee = cliutils.DerefString(issue.Assignee.Slug)
			}
			priority := cliutils.DerefString(issue.Priority)
			updatedAtStr := cliutils.DerefString(issue.UpdatedAt)
			updatedAtFmt := cliutils.FormatRelativeTime(updatedAtStr)

			if len(title) > 50 {
				title = title[:47] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				issueID,
				title,
				status,
				assignee,
				priority,
				updatedAtFmt,
			)
		}
		return w.Flush()
	},
}

func init() {
	issueCmd.AddCommand(issueListCmd)
	issueListCmd.Flags().StringVarP(&issueListRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (Default: current repository)")
}
