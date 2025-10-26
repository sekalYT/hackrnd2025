// cmd/issue_create.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/api"
	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	issueCreateTitleFlag       string
	issueCreateDescriptionFlag string
	issueCreateRepoFlag        string
)

var issueCreateCmd = &cobra.Command{
	Use:   "create [flags]",
	Short: "Create a new Issue",
	Long:  `Creates a new issue in the SourceCraft repository.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		var orgSlug, repoSlug string
		var err error

		if issueCreateRepoFlag != "" {
			parts := strings.SplitN(issueCreateRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", issueCreateRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			fmt.Println("Defining a repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("could not identify the repository from git remote and 'organization' is not set in the config. Use the --repo <org>/<repo>")
				}
				return fmt.Errorf("failed to identify repository slug from git remote. Use the --repo <org>/<repo>")
			}
		}
		fmt.Printf("Creating an issue in the repository: %s/%s\n", orgSlug, repoSlug)

		title := issueCreateTitleFlag
		if title == "" {
			title, err = promptForInput("Title", "")
			if err != nil {
				return err
			}
			if title == "" {
				return fmt.Errorf("the title cannot be empty")
			}
		}

		description := issueCreateDescriptionFlag
		if description == "" {
			description, err = promptForInput("Description (optional, Enter to skip)", "")
			if err != nil {
				return err
			}
		}

		apiBody := api.CreateIssueBody{
			Title:       title,
			Description: description,
		}

		fmt.Println("Creating Issue...")
		createdIssue, err := apiClient.CreateIssue(orgSlug, repoSlug, apiBody)
		if err != nil {
			return err
		}

		fmt.Println("\nThe issue has been successfully created!")
		fmt.Printf("ID/Slug:    %s\n", cliutils.DerefString(createdIssue.Slug))
		fmt.Printf("Title:  %s\n", cliutils.DerefString(createdIssue.Title))
		if createdIssue.Status != nil {
			fmt.Printf("Status:     %s\n", cliutils.DerefString(createdIssue.Status.Name))
		}

		webURL := fmt.Sprintf("https://sourcecraft.dev/%s/%s/issues/%s", orgSlug, repoSlug, cliutils.DerefString(createdIssue.Slug))
		fmt.Printf("View: %s\n", webURL)

		return nil
	},
}

func init() {
	issueCmd.AddCommand(issueCreateCmd)

	issueCreateCmd.Flags().StringVarP(&issueCreateTitleFlag, "title", "t", "", "Issue Title")
	issueCreateCmd.Flags().StringVarP(&issueCreateDescriptionFlag, "description", "d", "", "Issue Description")
	issueCreateCmd.Flags().StringVarP(&issueCreateRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (Default: Current repository)")
}
