// cmd/repo_view.go
package cmd

import (
	"fmt"
	"time"

	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var repoViewCmd = &cobra.Command{
	Use:   "view <repo_slug>",
	Short: "View information about a specific repository",
	Long: `Displays detailed information about a repository within the organization specified in config.yaml.
Example: src repo view my-awesome-repo`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoSlug := args[0]

		orgSlug := viper.GetString("organization")
		if orgSlug == "" {
			return fmt.Errorf("error: 'organization' slug not found in config.yaml or SOURCECRAFT_ORGANIZATION env var")
		}

		fmt.Printf("Fetching details for repository '%s/%s'...\n", orgSlug, repoSlug)

		repo, err := apiClient.GetRepository(orgSlug, repoSlug)
		if err != nil {
			return err
		}

		fmt.Println("--- Repository Details ---")
		fmt.Printf("Name:        %s\n", cliutils.DerefString(repo.Name))
		fmt.Printf("Slug:        %s\n", cliutils.DerefString(repo.Slug))
		fmt.Printf("Full Path:   %s/%s\n", orgSlug, cliutils.DerefString(repo.Slug))
		fmt.Printf("Visibility:  %s\n", cliutils.DerefString(repo.Visibility))
		fmt.Printf("Description: %s\n", cliutils.DerefString(repo.Description))
		fmt.Printf("Default Br:  %s\n", cliutils.DerefString(repo.DefaultBranch))
		fmt.Printf("Is Empty:    %t\n", cliutils.DerefBool(repo.IsEmpty))

		if repo.CloneURL != nil {
			fmt.Printf("SSH URL:     %s\n", cliutils.DerefString(repo.CloneURL.SSH))
			fmt.Printf("HTTPS URL:   %s\n", cliutils.DerefString(repo.CloneURL.HTTPS))
		}

		if repo.Owner != nil {

			fmt.Printf("Owner Slug:  %s\n", cliutils.DerefString(repo.Owner.Slug))
		}

		if repo.LastUpdated != nil {
			t, err := time.Parse(time.RFC3339Nano, *repo.LastUpdated)
			if err != nil {
				t, err = time.Parse(time.RFC3339, *repo.LastUpdated)
			}

			if err == nil {
				fmt.Printf("Last Update: %s\n", t.Local().Format("2006-01-02 15:04:05 MST"))
			} else {
				fmt.Printf("Last Update: %s (raw, parse error: %v)\n", *repo.LastUpdated, err)
			}
		}

		if repo.Language != nil {
			fmt.Printf("Language:    %s\n", cliutils.DerefString(repo.Language.Name))
		}

		return nil
	},
}

func init() {
	repoCmd.AddCommand(repoViewCmd)
}
