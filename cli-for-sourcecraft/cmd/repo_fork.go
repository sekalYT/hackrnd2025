// cmd/repo_fork.go
package cmd

import (
	"fmt"
	"strings"

	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	forkTargetOrgFlag         string
	forkRepoNameFlag          string
	forkDefaultBranchOnlyFlag bool
)

var repoForkCmd = &cobra.Command{
	Use:   "fork <source_repository> [flags]",
	Short: "Create a fork of a repository",
	Long: `Creates a fork (a personal copy) of another repository.
The fork will be created in the organization specified in your config.yaml unless overridden with --org.

<source_repository> format: <source_org_slug>/<source_repo_slug>
Example: src repo fork source-org/original-repo --name my-copy-slug`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceRepoIdentifier := args[0]

		parts := strings.SplitN(sourceRepoIdentifier, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid format for <source_repository>: '%s'. Expected format: <source_org_slug>/<source_repo_slug>", sourceRepoIdentifier)
		}
		sourceOrgSlug := parts[0]
		sourceRepoSlug := parts[1]

		targetOrgSlug := forkTargetOrgFlag
		if targetOrgSlug == "" {
			targetOrgSlug = viper.GetString("organization")
			if targetOrgSlug == "" {
				return fmt.Errorf("error: target organization not specified. Use --org flag or set 'organization' in config.yaml")
			}
			fmt.Printf("Forking into your organization '%s' (from config)...\n", targetOrgSlug)
		} else {
			fmt.Printf("Forking into specified organization '%s'...\n", targetOrgSlug)
		}

		newRepoSlug := forkRepoNameFlag

		targetRepoName := sourceRepoSlug
		if newRepoSlug != "" {
			targetRepoName = newRepoSlug
		}

		fmt.Printf("Forking '%s/%s' to '%s/%s'...\n", sourceOrgSlug, sourceRepoSlug, targetOrgSlug, targetRepoName)
		if forkDefaultBranchOnlyFlag {
			fmt.Println("Only copying the default branch.")
		}

		forkedRepo, err := apiClient.ForkRepository(sourceOrgSlug, sourceRepoSlug, targetOrgSlug, newRepoSlug, forkDefaultBranchOnlyFlag)
		if err != nil {
			return err
		}

		createdName := cliutils.DerefString(forkedRepo.Name)
		createdSlug := cliutils.DerefString(forkedRepo.Slug)
		sshUrl := ""
		if forkedRepo.CloneURL != nil && forkedRepo.CloneURL.SSH != nil {
			sshUrl = *forkedRepo.CloneURL.SSH
		}
		finalOrgSlug := targetOrgSlug
		if forkedRepo.Owner != nil && forkedRepo.Owner.Slug != nil {

			finalOrgSlug = *forkedRepo.Owner.Slug
		}

		fmt.Printf("Successfully forked repository to '%s/%s'\n", finalOrgSlug, createdName)
		fmt.Printf("New repository slug: %s\n", createdSlug)
		fmt.Println("SSH Clone URL:", sshUrl)

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoForkCmd)
	repoForkCmd.Flags().StringVar(&forkTargetOrgFlag, "org", "", "Organization slug to fork into (defaults to organization in config.yaml)")
	repoForkCmd.Flags().StringVar(&forkRepoNameFlag, "name", "", "Slug for the new forked repository (defaults to the original repository slug)") // Clarified help text
	repoForkCmd.Flags().BoolVar(&forkDefaultBranchOnlyFlag, "default-branch-only", false, "Only copy the default branch")
}
