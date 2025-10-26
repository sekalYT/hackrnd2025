// cmd/workflow_run.go
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
	wfRunRepoFlag             string
	wfRunRevisionFlag         string
	wfRunWorkflowRevisionFlag string
)

var workflowRunCmd = &cobra.Command{
	Use:   "run <workflow_name> [flags]",
	Short: "Run workflow by name",
	Long: `"Triggers a CI/CD workflow by its name (e.g., 'main' or 'build')."

"<workflow_name> is the name defined in the CI configuration, not the ID."

"Flags:"

--revision (-r): The branch, tag, or SHA on which to run the workflow (default: the repository's default branch).
--workflow-revision: The branch, tag, or SHA from which to fetch the workflow YML file (default: the same as --revision).
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowName := args[0]
		var orgSlug, repoSlug string
		var err error

		if wfRunRepoFlag != "" {
			parts := strings.SplitN(wfRunRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Coming soon: /", wfRunRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			fmt.Println("Defining a repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("could not identify the repository from git remote and 'organization' is not set in the config. Use the -- repo flag")
				}
				return fmt.Errorf("failed to identify the repository slug from Git Remote. Use the -- repo flag")
			}
		}

		apiBody := api.RunCIBody{
			WorkflowSlug:     workflowName,
			Revision:         wfRunRevisionFlag,
			WorkflowRevision: wfRunWorkflowRevisionFlag,
		}

		fmt.Printf("Run workflow '%s' в %s/%s", workflowName, orgSlug, repoSlug)
		if wfRunRevisionFlag != "" {
			fmt.Printf(" (on the branch/revision: %s)", wfRunRevisionFlag)
		}
		fmt.Println("...")

		runResponse, err := apiClient.RunWorkflow(orgSlug, repoSlug, workflowName, apiBody)
		if err != nil {
			return err
		}

		fmt.Println("\nWorkflow has been successfully launched!")
		fmt.Printf("Trigger status: %s\n", cliutils.DerefString(runResponse.TriggerStatus))
		fmt.Printf("Flux ID: %s\n", cliutils.DerefString(runResponse.FluxID))

		return nil
	},
}

func init() {
	workflowCmd.AddCommand(workflowRunCmd)
	workflowRunCmd.Flags().StringVarP(&wfRunRepoFlag, "repo", "R", "", "Specify repository in / format (default: current repository)")
	workflowRunCmd.Flags().StringVarP(&wfRunRevisionFlag, "revision", "r", "", "Branch, tag, or SHA to run (default: default-branch)")
	workflowRunCmd.Flags().StringVar(&wfRunWorkflowRevisionFlag, "workflow-revision", "", "Branch, tag, or SHA where to get the YML file from (default: same as --revision)")
}
