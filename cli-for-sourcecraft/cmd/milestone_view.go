// cmd/milestone_view.go
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
	milestoneViewRepoFlag string
)

var milestoneViewCmd = &cobra.Command{
	Use:   "view <milestone_slug>",
	Short: "View Milestone Details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		milestoneSlug := args[0]
		var orgSlug, repoSlug string
		var err error

		if milestoneViewRepoFlag != "" {
			parts := strings.SplitN(milestoneViewRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", milestoneViewRepoFlag)
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
		fmt.Printf("Milestone Request '%s' в %s/%s\n", milestoneSlug, orgSlug, repoSlug)

		ms, err := apiClient.GetMilestone(orgSlug, repoSlug, milestoneSlug)
		if err != nil {
			return err
		}

		fmt.Printf("\n--- %s ---\n", cliutils.DerefString(ms.Name))
		fmt.Printf("ID/Slug:     %s\n", cliutils.DerefString(ms.Slug))
		fmt.Printf("Status:      %s\n", cliutils.DerefString(ms.Status))

		if ms.Author != nil {
			fmt.Printf("Author:       %s\n", cliutils.DerefString(ms.Author.Slug))
		}

		fmt.Printf("Beginning:      %s\n", formatMilestoneDate(ms.StartDate))
		fmt.Printf("End:   %s\n", formatMilestoneDate(ms.Deadline))

		if ms.UpdatedAt != nil {
			updatedAtStr := cliutils.DerefString(ms.UpdatedAt)
			t, parseErr := time.Parse(time.RFC3339Nano, updatedAtStr)
			if parseErr != nil {
				t, parseErr = time.Parse(time.RFC3339, updatedAtStr)
			}
			if parseErr == nil {
				fmt.Printf("Updated:   %s\n", t.Local().Format("2006-01-02 15:04:05 MST"))
			}
		}

		webURL := fmt.Sprintf("https://sourcecraft.dev/%s/%s/milestones/%s", orgSlug, repoSlug, cliutils.DerefString(ms.Slug))
		fmt.Printf("View:  %s\n", webURL)

		fmt.Println("\n--- Description ---")
		fmt.Println(cliutils.DerefString(ms.Description))
		fmt.Println("------------------")

		return nil
	},
}

func init() {
	milestoneCmd.AddCommand(milestoneViewCmd)
	milestoneViewCmd.Flags().StringVarP(&milestoneViewRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (Default: current repository)")
}
