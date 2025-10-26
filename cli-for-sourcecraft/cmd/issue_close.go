// cmd/issue_close.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/api"
	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
)

var (
	issueCloseRepoFlag string
)

var issueCloseCmd = &cobra.Command{
	Use:   "close <issue_id_or_slug>",
	Short: "Close Issue",
	Long:  `Sets the status of the issue to 'closed'.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueSlug := args[0]
		var orgSlug, repoSlug string
		var err error

		if issueCloseRepoFlag != "" {
			parts := strings.SplitN(issueCloseRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", issueCloseRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("the repository could not be determined. Use the --repo <org>/<repo>")
			}
		}

		statusClosed := "closed"
		body := api.UpdateIssueBody{
			StatusSlug: &statusClosed,
		}

		fmt.Printf("Closing a issue #%s в %s/%s...\n", issueSlug, orgSlug, repoSlug)

		updatedIssue, err := apiClient.UpdateIssue(orgSlug, repoSlug, issueSlug, body)
		if err != nil {
			return err
		}

		fmt.Println("\nThe issue was successfully closed!")
		fmt.Printf("ID/Slug:    %s\n", cliutils.DerefString(updatedIssue.Slug))
		if updatedIssue.Status != nil {
			fmt.Printf("New status: %s\n", cliutils.DerefString(updatedIssue.Status.Name))
		}

		return nil
	},
}

func init() {
	issueCmd.AddCommand(issueCloseCmd)
	issueCloseCmd.Flags().StringVarP(&issueCloseRepoFlag, "repo", "R", "", "Specify repository <org>/<repo>")
}
