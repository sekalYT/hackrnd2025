// cmd/repo_list.go
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "View a list of your organization's repositories",
	Long:  `Shows a list of repositories owned by the organization specified in config.yaml.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		orgSlug := viper.GetString("organization")
		if orgSlug == "" {
			return fmt.Errorf("error: 'organization' not found in config.yaml.")
		}

		fmt.Printf("Request repositories for your organization '%s'...\n", orgSlug)

		repos, err := apiClient.ListRepositories(orgSlug)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories found.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSLUG\tVISIBILITY\tDESCRIPTION\tSSH URL")
		fmt.Fprintln(w, "----\t----\t----------\t-----------\t-------")
		for _, repo := range repos {
			name := cliutils.DerefString(repo.Name)
			slug := cliutils.DerefString(repo.Slug)
			visibility := cliutils.DerefString(repo.Visibility)
			description := cliutils.DerefString(repo.Description)
			sshUrl := ""
			if repo.CloneURL != nil && repo.CloneURL.SSH != nil {
				sshUrl = *repo.CloneURL.SSH
			}
			if len(description) > 50 {
				description = description[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				name, slug, visibility, description, sshUrl,
			)
		}
		return w.Flush()
	},
}

func init() {
	repoCmd.AddCommand(repoListCmd)
}
