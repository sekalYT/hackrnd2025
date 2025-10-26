// cmd/milestone_list.go
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	milestoneListRepoFlag string
)

var milestoneListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "View the list of milestones in the repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		var orgSlug, repoSlug string
		var err error

		if milestoneListRepoFlag != "" {
			parts := strings.SplitN(milestoneListRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", milestoneListRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			fmt.Println("Defining a repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("couldn't identify repository from git remote and 'organization' is not set in config. Use the --repo <org>/<repo>")
				}
				return fmt.Errorf("failed to identify repository slug from git remote. Use the --repo <org>/<repo>")
			}
		}
		fmt.Printf("Querying milestones for a repository: %s/%s\n", orgSlug, repoSlug)

		milestones, err := apiClient.ListMilestonesForRepository(orgSlug, repoSlug)
		if err != nil {
			return err
		}

		if len(milestones) == 0 {
			fmt.Println("No milestones found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID (SLUG)\tNAME\tSTATUS\tSTART DATE\tDEADLINE")
		fmt.Fprintln(w, "---------\t----\t------\t----------\t--------")

		for _, ms := range milestones {
			slug := cliutils.DerefString(ms.Slug)
			name := cliutils.DerefString(ms.Name)
			status := cliutils.DerefString(ms.Status)

			startDate := formatMilestoneDate(ms.StartDate)
			deadline := formatMilestoneDate(ms.Deadline)

			if len(name) > 40 {
				name = name[:37] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				slug,
				name,
				status,
				startDate,
				deadline,
			)
		}
		return w.Flush()
	},
}

func formatMilestoneDate(ts *string) string {
	if ts == nil || *ts == "" {
		return "-"
	}

	t, err := time.Parse(time.RFC3339Nano, *ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339, *ts)
		if err != nil {
			return *ts
		}
	}
	return t.Local().Format("2006-01-02")
}

func init() {
	milestoneCmd.AddCommand(milestoneListCmd)
	milestoneListCmd.Flags().StringVarP(&milestoneListRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (Default: current repository)")
}
