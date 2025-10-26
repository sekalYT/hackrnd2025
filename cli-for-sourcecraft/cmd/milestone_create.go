// cmd/milestone_create.go
package cmd

import (
	"fmt"
	"strings"
	"time"

	"cli-for-sourcecraft/internal/api"
	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	msCreateRepoFlag     string
	msCreateNameFlag     string
	msCreateSlugFlag     string
	msCreateDescFlag     string
	msCreateStartFlag    string
	msCreateDeadlineFlag string
)

var milestoneCreateCmd = &cobra.Command{
	Use:   "create [flags]",
	Short: "Create a new Milestone",
	Long: `Creates a new milestone in the repository.
Dates (--start, --deadline) must be specified in the format YYYY-MM-DD.

Example: src milestone create --name "Release 1.0" --deadline "2025-12-31"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		var orgSlug, repoSlug string
		var err error

		if msCreateNameFlag == "" {
			return fmt.Errorf("flag --name (milestone name) is mandatory")
		}

		if msCreateRepoFlag != "" {
			parts := strings.SplitN(msCreateRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", msCreateRepoFlag)
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
		fmt.Printf("Create a milestone in a repository: %s/%s\n", orgSlug, repoSlug)

		apiBody := api.CreateMilestoneBody{
			Name:        msCreateNameFlag,
			Slug:        msCreateSlugFlag,
			Description: msCreateDescFlag,
		}

		if msCreateStartFlag != "" {
			apiBody.StartDate, err = parseDateToRFC3339(msCreateStartFlag)
			if err != nil {
				return fmt.Errorf("invalid format --start-date: %w", err)
			}
		}
		if msCreateDeadlineFlag != "" {
			apiBody.Deadline, err = parseDateToRFC3339(msCreateDeadlineFlag)
			if err != nil {
				return fmt.Errorf("invalid format --deadline: %w", err)
			}
		}

		fmt.Println("Creating milestone...")
		createdMS, err := apiClient.CreateMilestone(orgSlug, repoSlug, apiBody)
		if err != nil {
			return err
		}

		fmt.Println("\nMilestone successfully created!")
		fmt.Printf("ID/Slug:    %s\n", cliutils.DerefString(createdMS.Slug))
		fmt.Printf("Title:   %s\n", cliutils.DerefString(createdMS.Name))
		fmt.Printf("Status:     %s\n", cliutils.DerefString(createdMS.Status))

		webURL := fmt.Sprintf("https://sourcecraft.dev/%s/%s/milestones/%s", orgSlug, repoSlug, cliutils.DerefString(createdMS.Slug))
		fmt.Printf("View: %s\n", webURL)

		return nil
	},
}

func parseDateToRFC3339(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("expected format YYYY-MM-DD, received '%s'", dateStr)
	}
	return t.UTC().Format(time.RFC3339), nil
}

func init() {
	milestoneCmd.AddCommand(milestoneCreateCmd)

	milestoneCreateCmd.Flags().StringVarP(&msCreateRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (Default: current repository)")
	milestoneCreateCmd.Flags().StringVarP(&msCreateNameFlag, "name", "n", "", "Milestone Name (Required)")
	milestoneCreateCmd.Flags().StringVarP(&msCreateDescFlag, "description", "d", "", "Milestone Description")
	milestoneCreateCmd.Flags().StringVar(&msCreateSlugFlag, "slug", "", "Slug (URL) milestones (generated from the name if not specified)")
	milestoneCreateCmd.Flags().StringVar(&msCreateStartFlag, "start-date", "", "Start Date (YYYY-MM-DD)")
	milestoneCreateCmd.Flags().StringVar(&msCreateDeadlineFlag, "deadline", "", "End Date (YYYY-MM-DD)")

	milestoneCreateCmd.MarkFlagRequired("name")
}
