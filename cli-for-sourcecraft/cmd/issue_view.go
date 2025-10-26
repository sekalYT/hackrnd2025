// cmd/issue_view.go
package cmd

import (
	"fmt"
	"strings"
	"time"

	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
)

var (
	issueViewRepoFlag string
)

var issueViewCmd = &cobra.Command{
	Use:   "view <issue_id_or_slug>",
	Short: "View detailed issue information",
	Long:  `Shows detailed information about the issue.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueSlug := args[0]
		var orgSlug, repoSlug string
		var err error

		if issueViewRepoFlag != "" {
			parts := strings.SplitN(issueViewRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", issueViewRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			fmt.Println("Defining a repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("the repository could not be determined. Use the --repo <org>/<repo>")
			}
		}
		fmt.Printf("Issue request #%s в %s/%s\n", issueSlug, orgSlug, repoSlug)

		issue, err := apiClient.GetIssue(orgSlug, repoSlug, issueSlug)
		if err != nil {
			return err
		}

		fmt.Printf("\n--- %s ---\n", cliutils.DerefString(issue.Title))
		fmt.Printf("ID/Slug:    %s\n", cliutils.DerefString(issue.Slug))

		if issue.Status != nil {
			fmt.Printf("Status:     %s (Тип: %s)\n", cliutils.DerefString(issue.Status.Name), cliutils.DerefString(issue.Status.StatusType))
		}
		if issue.Author != nil {
			fmt.Printf("Author:      %s\n", cliutils.DerefString(issue.Author.Slug))
		}
		if issue.Assignee != nil {
			fmt.Printf("Perform:%s\n", cliutils.DerefString(issue.Assignee.Slug))
		}
		fmt.Printf("Priority:  %s\n", cliutils.DerefString(issue.Priority))

		if issue.UpdatedAt != nil {
			updatedAtStr := cliutils.DerefString(issue.UpdatedAt)
			t, parseErr := time.Parse(time.RFC3339Nano, updatedAtStr)
			if parseErr != nil {
				t, parseErr = time.Parse(time.RFC3339, updatedAtStr)
			}
			if parseErr == nil {
				fmt.Printf("Updated:  %s\n", t.Local().Format("2006-01-02 15:04:05 MST"))
			}
		}

		webURL := fmt.Sprintf("https://sourcecraft.dev/%s/%s/issues/%s", orgSlug, repoSlug, cliutils.DerefString(issue.Slug))
		fmt.Printf("View: %s\n", webURL)

		fmt.Println("\n--- Description ---")
		fmt.Println(cliutils.DerefString(issue.Description))
		fmt.Println("------------------")

		return nil
	},
}

func init() {
	issueCmd.AddCommand(issueViewCmd)
	issueViewCmd.Flags().StringVarP(&issueViewRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (Default: current repository)")
}
