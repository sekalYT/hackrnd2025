// cmd/workflow_status.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
)

var wfStatusRepoFlag string

var workflowStatusCmd = &cobra.Command{
	Use:   "status <run_slug> [flags]",
	Short: "View detailed status of CI/CD startup",
	Long:  `Shows the status and progress of execution for a specific CI/CD (Run Slug) run`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runSlug := args[0]
		var orgSlug, repoSlug string
		var err error

		if wfStatusRepoFlag != "" {
			parts := strings.SplitN(wfStatusRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", wfStatusRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("the repository could not be determined. Use the --repo <org>/<repo>")
			}
		}

		fmt.Printf("Request launch status '%s' в %s/%s\n", runSlug, orgSlug, repoSlug)

		run, err := apiClient.GetRunStatus(orgSlug, repoSlug, runSlug)
		if err != nil {
			return err
		}

		fmt.Printf("\n--- CI/CD Run: %s ---\n", cliutils.DerefString(run.Slug))
		fmt.Printf("General status: %s\n", cliutils.DerefString(run.Status))
		fmt.Printf("Launched: %s\n", cliutils.DerefString(run.CreatedAt))
		fmt.Printf("Updated: %s\n", cliutils.DerefString(run.UpdatedAt))
		fmt.Println("--------------------")

		for i, wf := range run.WorkflowRuns {
			wfStatus := cliutils.DerefString(wf.Status)
			wfSlug := cliutils.DerefString(wf.WorkflowSlug)

			fmt.Printf("  Workflow #%d: %s (Status: %s)\n", i+1, wfSlug, wfStatus)
			for j, task := range wf.TaskRuns {
				taskStatus := cliutils.DerefString(task.Status)
				taskSlug := cliutils.DerefString(task.TaskSlug)
				fmt.Printf("    - Task #%d: %s (Status: %s)\n", j+1, taskSlug, taskStatus)

				for _, cube := range task.CubeRuns {
					cubeSlug := cliutils.DerefString(cube.CubeSlug)
					fmt.Printf("      - Cube: %s (Status: %s) -> Logs: src wf logs %s %s %s %s\n",
						cubeSlug, cliutils.DerefString(cube.Status),
						runSlug, wfSlug, taskSlug, cubeSlug)
				}
			}
		}

		return nil
	},
}

func init() {
	workflowCmd.AddCommand(workflowStatusCmd)
	workflowStatusCmd.Flags().StringVarP(&wfStatusRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo>")
}
