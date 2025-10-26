// cmd/repo_clone.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cloneUseHTTPS bool
)

var repoCloneCmd = &cobra.Command{
	Use:   "clone <repository> [directory]",
	Short: "Clone a repository from SourceCraft",
	Long: `Clones a repository from SourceCraft into a new directory.

<repository> can be one of the following formats:
  - <repo_slug>         (e.g., my-awesome-repo) - Clones from the organization in your config.
  - <org_slug>/<repo_slug> (e.g., organization-sekal01/my-awesome-repo)
  - Full HTTPS URL
  - Full SSH URL

[directory] is optional. If not provided, the repository slug will be used.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoIdentifier := args[0]
		targetDirectory := ""
		if len(args) == 2 {
			targetDirectory = args[1]
		}

		var cloneURL string
		var repoSlugForDir string

		if strings.HasPrefix(repoIdentifier, "https://") || strings.HasPrefix(repoIdentifier, "git@") {
			fmt.Println("Using provided URL:", repoIdentifier)
			cloneURL = repoIdentifier
			parts := strings.Split(strings.TrimSuffix(repoIdentifier, ".git"), "/")
			if len(parts) > 0 {
				repoSlugForDir = parts[len(parts)-1]
			}

		} else {
			orgSlug := viper.GetString("organization")
			repoSlug := repoIdentifier

			if strings.Contains(repoIdentifier, "/") {
				parts := strings.SplitN(repoIdentifier, "/", 2)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return fmt.Errorf("invalid format for <org_slug>/<repo_slug>: %s", repoIdentifier)
				}
				orgSlug = parts[0]
				repoSlug = parts[1]
				fmt.Printf("Cloning repository '%s' from organization '%s'...\n", repoSlug, orgSlug)
			} else {
				if orgSlug == "" {
					return fmt.Errorf("error: 'organization' slug not found in config.yaml. Provide full slug <org>/<repo> or set organization in config")
				}
				fmt.Printf("Cloning repository '%s' from organization '%s' (from config)...\n", repoSlug, orgSlug)
			}

			if repoSlug == "" {
				return fmt.Errorf("repository slug cannot be empty")
			}
			repoSlugForDir = repoSlug

			repo, err := apiClient.GetRepository(orgSlug, repoSlug)
			if err != nil {
				return err
			}

			if repo.CloneURL == nil {
				return fmt.Errorf("API did not provide clone URLs for repository %s/%s", orgSlug, repoSlug)
			}
			if cloneUseHTTPS {
				if repo.CloneURL.HTTPS != nil && *repo.CloneURL.HTTPS != "" {
					cloneURL = *repo.CloneURL.HTTPS
					fmt.Println("Using HTTPS clone URL.")
				} else {
					return fmt.Errorf("HTTPS clone URL not available for this repository")
				}
			} else {
				if repo.CloneURL.SSH != nil && *repo.CloneURL.SSH != "" {
					cloneURL = *repo.CloneURL.SSH
					fmt.Println("Using SSH clone URL.")
				} else if repo.CloneURL.HTTPS != nil && *repo.CloneURL.HTTPS != "" {
					cloneURL = *repo.CloneURL.HTTPS
					fmt.Println("SSH URL not available, falling back to HTTPS clone URL.")
				} else {
					return fmt.Errorf("no suitable clone URL (SSH or HTTPS) available for this repository")
				}
			}
		}

		if cloneURL == "" {
			return fmt.Errorf("could not determine clone URL")
		}

		if targetDirectory == "" {
			if repoSlugForDir != "" {
				targetDirectory = repoSlugForDir
			} else {
				return fmt.Errorf("could not determine target directory name, please specify it as the second argument")
			}
		}

		fmt.Printf("Cloning into '%s'...\n", targetDirectory)
		gitArgs := []string{"clone", cloneURL, targetDirectory}
		gitCmd := exec.Command("git", gitArgs...)

		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		err := gitCmd.Run()
		if err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}

		fmt.Println("\nRepository cloned successfully.")
		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoCloneCmd)
	repoCloneCmd.Flags().BoolVar(&cloneUseHTTPS, "https", false, "Use HTTPS URL for cloning instead of SSH")

}
