// cmd/workflow_logs.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/git"

	"github.com/spf13/cobra"
)

var wfLogsRepoFlag string

var workflowLogsCmd = &cobra.Command{
	Use:   "logs <run_slug> <workflow_slug> <task_slug> <cube_slug> [flags]",
	Short: "View CI/CD Cube Logs",
	Long: `Displays logs for a specific cube (the smallest unit of execution) in CI/CD.
ou need to know all 4 slugs. You can find them in the output of 'src workflow status <run_slug>'.`,
	Args: cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		runSlug := args[0]
		workflowSlug := args[1]
		taskSlug := args[2]
		cubeSlug := args[3]
		var orgSlug, repoSlug string
		var err error

		if wfLogsRepoFlag != "" {
			parts := strings.SplitN(wfLogsRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", wfLogsRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("the repository could not be determined. Use the --repo <org>/<repo>")
			}
		}

		fmt.Printf("Log request for %s/%s | Run: %s, WF: %s, Task: %s, Cube: %s\n",
			orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug)

		logs, err := apiClient.GetLogs(orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug)
		if err != nil {
			return err
		}

		fmt.Println("\n--- LOGS ---")
		fmt.Println(logs)
		fmt.Println("------------")

		return nil
	},
}

func init() {
	workflowCmd.AddCommand(workflowLogsCmd)
	workflowLogsCmd.Flags().StringVarP(&wfLogsRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo>")
}
