// cmd/pr_list.go
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

var (
	prListRepoFlag string
)

var prListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "List pull requests in a repository",
	Long: `Lists pull requests for a specified repository.
If no repository is specified with --repo, it uses the current repository based on git remotes.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		var orgSlug, repoSlug string
		var err error

		if prListRepoFlag != "" {
			parts := strings.SplitN(prListRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid format for --repo flag: '%s'. Expected format: <org_slug>/<repo_slug>", prListRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
			fmt.Printf("Listing pull requests for specified repository: %s/%s\n", orgSlug, repoSlug)
		} else {
			fmt.Println("Attempting to detect repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				orgSlug = viper.GetString("organization")
				if orgSlug == "" {
					return fmt.Errorf("could not detect repository from git remote and 'organization' not set in config. Use --repo <org>/<repo> flag or run from within a repository")
				}
				return fmt.Errorf("could not detect repository slug from git remote. Use --repo <org>/<repo> flag or run from within a repository")

			}
			fmt.Printf("Detected repository: %s/%s\n", orgSlug, repoSlug)
		}

		fmt.Printf("Fetching pull requests for %s/%s...\n", orgSlug, repoSlug)
		prs, err := apiClient.ListPullRequests(orgSlug, repoSlug)
		if err != nil {
			return err
		}

		if len(prs) == 0 {
			fmt.Println("No pull requests found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSOURCE -> TARGET\tSTATUS\tAUTHOR\tUPDATED")
		fmt.Fprintln(w, "--\t-----\t----------------\t------\t------\t-------")

		for _, pr := range prs {
			prID := cliutils.DerefString(pr.Slug)
			title := cliutils.DerefString(pr.Title)
			source := cliutils.DerefString(pr.SourceBranch)
			target := cliutils.DerefString(pr.TargetBranch)
			status := cliutils.DerefString(pr.Status)
			author := ""
			if pr.Author != nil {
				author = cliutils.DerefString(pr.Author.Slug)
			}
			updatedAtStr := cliutils.DerefString(pr.UpdatedAt)
			updatedAtFmt := cliutils.FormatRelativeTime(updatedAtStr)

			if len(title) > 50 {
				title = title[:47] + "..."
			}
			branchInfo := fmt.Sprintf("%s -> %s", source, target)

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				prID,
				title,
				branchInfo,
				status,
				author,
				updatedAtFmt,
			)
		}
		return w.Flush()
	},
}

func init() {
	prCmd.AddCommand(prListCmd)
	prListCmd.Flags().StringVarP(&prListRepoFlag, "repo", "R", "", "Specify repository in <org>/<repo> format")
}
