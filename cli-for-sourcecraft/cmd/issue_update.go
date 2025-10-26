// cmd/issue_update.go
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
	issueUpdateRepoFlag        string
	issueUpdateTitleFlag       string
	issueUpdateDescriptionFlag string
	issueUpdateStatusFlag      string
	issueUpdatePriorityFlag    string
	issueUpdateAssigneeFlag    string
)

var issueUpdateCmd = &cobra.Command{
	Use:   "update <issue_id_or_slug> [flags]",
	Short: "Update Issue",
	Long: `Updates task fields such as title, description, status, etc.
Example: src issue update 12 --title "New title" --status "inProgress"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueSlug := args[0]
		var orgSlug, repoSlug string
		var err error

		if issueUpdateRepoFlag != "" {
			parts := strings.SplitN(issueUpdateRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", issueUpdateRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("the repository could not be determined. Use the --repo <org>/<repo>")
			}
		}

		var body api.UpdateIssueBody
		hasChanges := false

		if cmd.Flags().Changed("title") {
			body.Title = &issueUpdateTitleFlag
			hasChanges = true
		}
		if cmd.Flags().Changed("description") {
			body.Description = &issueUpdateDescriptionFlag
			hasChanges = true
		}
		if cmd.Flags().Changed("status") {
			body.StatusSlug = &issueUpdateStatusFlag
			hasChanges = true
		}
		if cmd.Flags().Changed("priority") {
			body.Priority = &issueUpdatePriorityFlag
			hasChanges = true
		}
		if cmd.Flags().Changed("assignee") {
			body.AssigneeID = &issueUpdateAssigneeFlag
			hasChanges = true
		}

		if !hasChanges {
			fmt.Println("No flags are specified for the update. Completion.")
			fmt.Println("Use the --title, --description, --status, --priority, --assignee.")
			return nil
		}

		fmt.Printf("Update issue #%s в %s/%s...\n", issueSlug, orgSlug, repoSlug)

		updatedIssue, err := apiClient.UpdateIssue(orgSlug, repoSlug, issueSlug, body)
		if err != nil {
			return err
		}

		fmt.Println("\nThe issue has been successfully updated!")
		fmt.Printf("ID/Slug:    %s\n", cliutils.DerefString(updatedIssue.Slug))
		fmt.Printf("Title:  %s\n", cliutils.DerefString(updatedIssue.Title))
		if updatedIssue.Status != nil {
			fmt.Printf("Status:     %s\n", cliutils.DerefString(updatedIssue.Status.Name))
		}
		if updatedIssue.Assignee != nil {
			fmt.Printf("Performer:%s\n", cliutils.DerefString(updatedIssue.Assignee.Slug))
		}

		return nil
	},
}

func init() {
	issueCmd.AddCommand(issueUpdateCmd)

	issueUpdateCmd.Flags().StringVarP(&issueUpdateRepoFlag, "repo", "R", "", "Specify a repository <org>/<repo>")
	issueUpdateCmd.Flags().StringVarP(&issueUpdateTitleFlag, "title", "t", "", "New issue title")
	issueUpdateCmd.Flags().StringVarP(&issueUpdateDescriptionFlag, "description", "d", "", "New issue description")
	issueUpdateCmd.Flags().StringVar(&issueUpdateStatusFlag, "status", "", "New status (slug): open, inProgress, closed, ...")
	issueUpdateCmd.Flags().StringVar(&issueUpdatePriorityFlag, "priority", "", "A new priority: trivial, minor, normal, ...")
	issueUpdateCmd.Flags().StringVarP(&issueUpdateAssigneeFlag, "assignee", "a", "", "ID of the user to whom the task is assigned ('' to remove)")
}
