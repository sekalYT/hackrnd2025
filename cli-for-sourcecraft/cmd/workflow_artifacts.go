// cmd/workflow_artifacts.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cli-for-sourcecraft/internal/git"

	"github.com/spf13/cobra"
)

var wfArtifactsRepoFlag string
var wfArtifactsOutputFlag string

var workflowArtifactsCmd = &cobra.Command{
	Use:   "artifacts <run_slug> <workflow_slug> <task_slug> <cube_slug> [flags]",
	Short: "Download a CI/CD cube artifact",
	Long: `Downloads a binary artifact associated with a specific CI/CD execution cube. Use the --output flag to specify the path to the file.
Use the --output flag to specify the path to the file.`,
	Args: cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		runSlug := args[0]
		workflowSlug := args[1]
		taskSlug := args[2]
		cubeSlug := args[3]
		var orgSlug, repoSlug string
		var err error

		if wfArtifactsRepoFlag != "" {
			parts := strings.SplitN(wfArtifactsRepoFlag, "/", 2)
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("the repository could not be determined. Use the --repo <org>/<repo>")
			}
		}

		outputFile := wfArtifactsOutputFlag
		if outputFile == "" {
			outputFile = fmt.Sprintf("%s-%s-%s-%s.artifact", runSlug, workflowSlug, taskSlug, cubeSlug)
		}

		fmt.Printf("Querying an artifact for %s/%s | Run: %s...\n", orgSlug, repoSlug, runSlug)

		data, err := apiClient.GetArtifacts(orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return fmt.Errorf("artifact Found but Empty")
		}

		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			return fmt.Errorf("directory could not be created: %w", err)
		}
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return fmt.Errorf("could not save the artifact to '%s': %w", outputFile, err)
		}

		fmt.Printf("\nThe artifact has been successfully saved: %s (size: %d bytes)\n", outputFile, len(data))

		return nil
	},
}

func init() {
	workflowCmd.AddCommand(workflowArtifactsCmd)
	workflowArtifactsCmd.Flags().StringVarP(&wfArtifactsRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo>")
	workflowArtifactsCmd.Flags().StringVarP(&wfArtifactsOutputFlag, "output", "o", "", "Path to save the artifact (default: <run_slug>-...artifact)")
}
