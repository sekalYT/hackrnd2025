// cmd/workflow_list.go
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

var wfListRepoFlag string

var workflowListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "View the list of CI/CD runs for a repository",
	Long:  `Shows a list of all CI/CD runs for the specified repository.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var orgSlug, repoSlug string
		var err error

		if wfListRepoFlag != "" {
			parts := strings.SplitN(wfListRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", wfListRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("could not identify repository from git remote and 'organization' not specified in the config. Use the --repo <org>/<repo>")
				}
				return fmt.Errorf("failed to identify repository slug from git remote. Use the --repo <org>/<repo>")
			}
		}
		fmt.Printf("Request CI/CD runs for: %s/%s\n", orgSlug, repoSlug)

		runs, err := apiClient.ListRuns(orgSlug, repoSlug)
		if err != nil {
			return err
		}

		if len(runs) == 0 {
			fmt.Println("No CI/CD runs found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID (SLUG)\tSTATUS\tWORKFLOWS\tUPDATED")
		fmt.Fprintln(w, "---------\t------\t---------\t-------")

		for _, run := range runs {
			slug := cliutils.DerefString(run.Slug)
			status := cliutils.DerefString(run.Status)

			wfCount := len(run.WorkflowRuns)
			wfStatus := fmt.Sprintf("%d wf", wfCount)

			updatedAtStr := cliutils.DerefString(run.UpdatedAt)
			updatedAtFmt := cliutils.FormatRelativeTime(updatedAtStr)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				slug,
				status,
				wfStatus,
				updatedAtFmt,
			)
		}
		return w.Flush()
	},
}

func init() {
	workflowCmd.AddCommand(workflowListCmd)
	workflowListCmd.Flags().StringVarP(&wfListRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo>")
}
