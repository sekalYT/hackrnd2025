// cmd/repo_create.go
package cmd

import (
	"fmt"
	"regexp"
	"strings"

	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	createDescriptionFlag string
	createSlugFlag        string
	createVisibilityFlag  string
)

var repoCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new repository in the organization",
	Long: `Creates a new repository within the organization specified in config.yaml.
Example: src repo create "My Awesome Project" -d "Description here" --slug my-awesome-project --visibility private`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]

		orgSlug := viper.GetString("organization")
		if orgSlug == "" {
			return fmt.Errorf("error: 'organization' slug not found in config.yaml or SOURCECRAFT_ORGANIZATION env var")
		}

		repoSlug := createSlugFlag
		if repoSlug == "" {
			repoSlug = generateSlug(repoName)
			fmt.Printf("Generated slug: %s (use --slug to override)\n", repoSlug)
		}

		visibility := createVisibilityFlag
		if visibility != "" && visibility != "public" && visibility != "private" && visibility != "internal" {
			return fmt.Errorf("invalid value for --visibility: '%s'. Allowed: public, internal, private", visibility)
		}

		fmt.Printf("Creating repository '%s/%s' in organization '%s'...\n", orgSlug, repoSlug, orgSlug)

		repo, err := apiClient.CreateRepository(orgSlug, repoName, repoSlug, createDescriptionFlag, visibility)
		if err != nil {
			return err
		}

		createdName := cliutils.DerefString(repo.Name)
		createdSlug := cliutils.DerefString(repo.Slug)
		sshUrl := ""
		if repo.CloneURL != nil && repo.CloneURL.SSH != nil {
			sshUrl = *repo.CloneURL.SSH
		}

		fmt.Printf("Successfully created repository '%s/%s'\n", orgSlug, createdName)
		fmt.Printf("Slug: %s\n", createdSlug)
		fmt.Println("SSH Clone URL:", sshUrl)

		return nil
	},
}

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9-]+`)
var slugMultipleHyphens = regexp.MustCompile(`-+`)

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugInvalidChars.ReplaceAllString(slug, "")
	slug = slugMultipleHyphens.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 256 {
		slug = slug[:256]
	}
	if slug == "" {
		return "repository"
	}
	return slug
}

func init() {
	repoCmd.AddCommand(repoCreateCmd)
	repoCreateCmd.Flags().StringVarP(&createDescriptionFlag, "description", "d", "", "Repository description")
	repoCreateCmd.Flags().StringVar(&createSlugFlag, "slug", "", "Repository slug (URL-friendly name, required by API, auto-generated if omitted)")
	repoCreateCmd.Flags().StringVar(&createVisibilityFlag, "visibility", "", "Repository visibility: public, internal, private (defaults to organization/server setting)")
}
