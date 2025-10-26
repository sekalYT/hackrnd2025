// cmd/pr_list.go
package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	// For formatting dates
	"cli-for-sourcecraft/internal/git"            // Import git for getting current repo
	cliutils "cli-for-sourcecraft/internal/utils" // Import utils

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flags for pr list
var (
	prListRepoFlag string // Flag to specify repo, e.g., --repo my-org/my-repo
	// Add flags for state (open, closed, etc.) later
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

		// Determine repository slug
		if prListRepoFlag != "" {
			// Use repo from flag
			parts := strings.SplitN(prListRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid format for --repo flag: '%s'. Expected format: <org_slug>/<repo_slug>", prListRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
			fmt.Printf("Listing pull requests for specified repository: %s/%s\n", orgSlug, repoSlug)
		} else {
			// Try to get from current git directory
			fmt.Println("Attempting to detect repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				// Fallback to reading organization from config, but we still need repo slug
				orgSlug = viper.GetString("organization") // Get org from config as fallback
				if orgSlug == "" {
					return fmt.Errorf("could not detect repository from git remote and 'organization' not set in config. Use --repo <org>/<repo> flag or run from within a repository")
				}
				// We cannot reliably guess the repoSlug if not in a git dir
				return fmt.Errorf("could not detect repository slug from git remote. Use --repo <org>/<repo> flag or run from within a repository")

			}
			fmt.Printf("Detected repository: %s/%s\n", orgSlug, repoSlug)
		}

		// Fetch PRs using the API client
		fmt.Printf("Fetching pull requests for %s/%s...\n", orgSlug, repoSlug)
		prs, err := apiClient.ListPullRequests(orgSlug, repoSlug) // Call the new API function
		if err != nil {
			return err // API errors will be shown
		}

		if len(prs) == 0 {
			fmt.Println("No pull requests found.")
			return nil
		}

		// Display PRs in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0) // Adjust padding
		// Use Slug field which likely holds the PR number/ID
		fmt.Fprintln(w, "ID\tTITLE\tSOURCE -> TARGET\tSTATUS\tAUTHOR\tUPDATED")
		fmt.Fprintln(w, "--\t-----\t----------------\t------\t------\t-------")

		for _, pr := range prs {
			prID := cliutils.DerefString(pr.Slug) // Use Slug as ID
			title := cliutils.DerefString(pr.Title)
			source := cliutils.DerefString(pr.SourceBranch)
			target := cliutils.DerefString(pr.TargetBranch)
			status := cliutils.DerefString(pr.Status)
			author := ""
			if pr.Author != nil {
				author = cliutils.DerefString(pr.Author.Slug) // Display author's slug
			}
			updatedAtStr := cliutils.DerefString(pr.UpdatedAt)
			updatedAtFmt := cliutils.FormatRelativeTime(updatedAtStr) // Use helper

			// Truncate long titles
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
	prCmd.AddCommand(prListCmd) // Add 'list' to 'pr'
	// Add --repo flag
	prListCmd.Flags().StringVarP(&prListRepoFlag, "repo", "R", "", "Specify repository in <org>/<repo> format")
	// TODO: Add flags like --state (open|closed|merged|all), --author, --assignee later
}
