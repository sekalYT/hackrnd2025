// cmd/pr_create.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"cli-for-sourcecraft/internal/api"
	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	prCreateTitleFlag      string
	prCreateBodyFlag       string
	prCreateBaseBranchFlag string
	prCreateHeadBranchFlag string
	prCreateRepoFlag       string
	prCreateReviewersFlag  []string
	prCreateDraftFlag      bool
)

var prCreateCmd = &cobra.Command{
	Use:   "create [flags]",
	Short: "Create a pull request",
	Long: `Creates a Pull Request on SourceCraft.

By default, it uses the current branch as the source (--head) and the repository's default branch (main/master) as the target (--base).
It will prompt for a Title and Description if they are not provided via flags.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		var orgSlug, repoSlug string
		var err error
		var repoInfo *api.Repo

		if prCreateRepoFlag != "" {
			parts := strings.SplitN(prCreateRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid flag format --repo: '%s'. Expected: <org>/<repo>", prCreateRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
			fmt.Printf("Target repository: %s/%s\n", orgSlug, repoSlug)
			fmt.Println("Get repository details...")
			repoInfo, err = apiClient.GetRepository(orgSlug, repoSlug)
			if err != nil {
				fmt.Printf("Note: Repository details could not be retrieved: %v\n", err)
			}
		} else {
			fmt.Println("Defining a repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("could not identify repository from git remote and 'organization' not specified in the config. Use the --repo <org>/<repo>")
				}
				return fmt.Errorf("failed to identify repository slug from git remote. Use the --repo <org>/<repo>")
			}
			fmt.Printf("Repository defined: %s/%s\n", orgSlug, repoSlug)
			fmt.Println("Get repository details...")
			repoInfo, err = apiClient.GetRepository(orgSlug, repoSlug)
			if err != nil {
				fmt.Printf("Warning: Repository details could not be retrieved: %v\n", err)
			}
		}

		headBranch := prCreateHeadBranchFlag
		if headBranch == "" {
			fmt.Println("Determining the current branch...")
			headBranch, err = git.GetCurrentBranchName()
			if err != nil {
				return fmt.Errorf("could not get the current branch: %w. Use the --head", err)
			}
			fmt.Printf("Using the current branch as the source branch: %s\n", headBranch)
		} else {
			fmt.Printf("Use the specified source branch: %s\n", headBranch)
		}

		baseBranch := prCreateBaseBranchFlag
		if baseBranch == "" {
			fmt.Println("Definition of the default branch...")
			baseBranch, err = git.GetDefaultBranchName(repoInfo, "origin")
			if err != nil {
				return fmt.Errorf("the default branch could not be determined: %w. Use the --base", err)
			}
			fmt.Printf("Use the default branch of the repository as the target branch: %s\n", baseBranch)
		} else {
			fmt.Printf("Use the specified target branch: %s\n", baseBranch)
		}

		if headBranch == baseBranch {
			return fmt.Errorf("the source branch ('%s') and the target branch ('%s') cannot be the same", headBranch, baseBranch)
		}

		title := prCreateTitleFlag
		if title == "" {
			commitTitle, _ := git.GetLastCommitTitle(headBranch)
			title, err = promptForInput("Title", commitTitle)
			if err != nil {
				return err
			}
			if title == "" {
				return fmt.Errorf("the title cannot be empty")
			}
		}

		body := prCreateBodyFlag
		if body == "" {
			commitBody, _ := git.GetCommitMessagesSinceBase(baseBranch, headBranch)
			body, err = promptForInput("Description (optional, Enter to skip)", commitBody)
			if err != nil {
				return err
			}
		}

		publishStatus := !prCreateDraftFlag
		apiBody := api.CreatePullRequestBody{
			Title:        title,
			SourceBranch: headBranch,
			TargetBranch: baseBranch,
			Description:  body,
			Publish:      publishStatus,
		}

		statusMsg := "Creating a pull request..."
		if prCreateDraftFlag {
			statusMsg = "Creating a draft Pull Request..."
		}
		fmt.Println(statusMsg)

		createdPR, err := apiClient.CreatePullRequest(orgSlug, repoSlug, apiBody)
		if err != nil {
			return err
		}

		fmt.Println("\nPull-Request has been successfully created!")
		fmt.Printf("ID/Slug:    %s\n", cliutils.DerefString(createdPR.Slug))
		fmt.Printf("Title:  %s\n", cliutils.DerefString(createdPR.Title))
		fmt.Printf("Status:     %s\n", cliutils.DerefString(createdPR.Status))
		fmt.Printf("From the branch:   %s\n", cliutils.DerefString(createdPR.SourceBranch))
		fmt.Printf("In the branch:    %s\n", cliutils.DerefString(createdPR.TargetBranch))
		webURL := fmt.Sprintf("https://sourcecraft.dev/%s/%s/pr/%s", orgSlug, repoSlug, cliutils.DerefString(createdPR.Slug))
		fmt.Printf("View: %s\n", webURL)

		return nil
	},
}

func promptForInput(prompt, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		cleanDefault := strings.ReplaceAll(strings.Split(defaultValue, "\n")[0], "\r", "")
		fmt.Printf("%s [%s]: ", prompt, cleanDefault)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func init() {
	prCmd.AddCommand(prCreateCmd)

	prCreateCmd.Flags().StringVarP(&prCreateTitleFlag, "title", "t", "", "Pull-Request Title")
	prCreateCmd.Flags().StringVarP(&prCreateBodyFlag, "body", "b", "", "Pull-Request Description")
	prCreateCmd.Flags().StringVarP(&prCreateBaseBranchFlag, "base", "B", "", "Target branch (where to measure) (default: default repository branch)")
	prCreateCmd.Flags().StringVarP(&prCreateHeadBranchFlag, "head", "H", "", "Source branch (where to freeze from) (default: current branch)")
	prCreateCmd.Flags().StringVarP(&prCreateRepoFlag, "repo", "R", "", "Specify a repository in the format <org>/<repo> (default: current repository)")
	prCreateCmd.Flags().BoolVarP(&prCreateDraftFlag, "draft", "d", false, "Create a Pull Request as a draft")

}
