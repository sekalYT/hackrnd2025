// cmd/repo_create.go
package cmd

import (
	"fmt"
	"regexp" // Import regexp for better slug generation
	"strings"

	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// No direct api import needed here if derefString is moved/duplicated
	// "cli-for-sourcecraft/internal/api"
)

// Flags for create command
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
	Args: cobra.ExactArgs(1), // Requires <name>
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]

		// Read organization slug from config/env
		orgSlug := viper.GetString("organization")
		if orgSlug == "" {
			return fmt.Errorf("error: 'organization' slug not found in config.yaml or SOURCECRAFT_ORGANIZATION env var")
		}

		// Determine repository slug: use flag or generate from name
		repoSlug := createSlugFlag
		if repoSlug == "" {
			repoSlug = generateSlug(repoName)
			fmt.Printf("Generated slug: %s (use --slug to override)\n", repoSlug)
		}

		// Determine visibility: use flag or let server default (by sending empty string)
		// Default is set on the flag itself in init()
		visibility := createVisibilityFlag
		// Validate if provided
		if visibility != "" && visibility != "public" && visibility != "private" && visibility != "internal" {
			return fmt.Errorf("invalid value for --visibility: '%s'. Allowed: public, internal, private", visibility)
		}

		fmt.Printf("Creating repository '%s/%s' in organization '%s'...\n", orgSlug, repoSlug, orgSlug)

		// apiClient is already initialized in rootCmd's PersistentPreRunE
		repo, err := apiClient.CreateRepository(orgSlug, repoName, repoSlug, createDescriptionFlag, visibility)
		if err != nil {
			return err // Error will include path like /orgs/{orgSlug}/repos
		}

		// Safely dereference pointer fields for output
		createdName := cliutils.DerefString(repo.Name)
		createdSlug := cliutils.DerefString(repo.Slug)
		sshUrl := ""
		if repo.CloneURL != nil && repo.CloneURL.SSH != nil {
			sshUrl = *repo.CloneURL.SSH
		}

		fmt.Printf("Successfully created repository '%s/%s'\n", orgSlug, createdName) // Use name returned by API
		fmt.Printf("Slug: %s\n", createdSlug)
		fmt.Println("SSH Clone URL:", sshUrl)
		// Optionally print HTTPS URL too
		// if repo.CloneURL != nil && repo.CloneURL.HTTPS != nil {
		// 	fmt.Println("HTTPS Clone URL:", *repo.CloneURL.HTTPS)
		// }
		return nil
	},
}

// Improved slug generation
var slugInvalidChars = regexp.MustCompile(`[^a-z0-9-]+`)
var slugMultipleHyphens = regexp.MustCompile(`-+`)

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")              // Replace spaces first
	slug = slugInvalidChars.ReplaceAllString(slug, "")     // Remove invalid chars
	slug = slugMultipleHyphens.ReplaceAllString(slug, "-") // Collapse multiple hyphens
	slug = strings.Trim(slug, "-")                         // Trim leading/trailing hyphens
	if len(slug) > 256 {                                   // Max length from Swagger
		slug = slug[:256]
	}
	// Handle empty slug after cleanup
	if slug == "" {
		// You might want a better default or return an error
		return "repository"
	}
	return slug
}

// Helper to safely dereference *string (duplicate from repo_list, move to utils?)
// func derefString(s *string) string {
// 	if s != nil { return *s }
// 	return ""
// }

func init() {
	repoCmd.AddCommand(repoCreateCmd)
	// Flags based on Swagger and TZ
	repoCreateCmd.Flags().StringVarP(&createDescriptionFlag, "description", "d", "", "Repository description")
	repoCreateCmd.Flags().StringVar(&createSlugFlag, "slug", "", "Repository slug (URL-friendly name, required by API, auto-generated if omitted)")
	// Let the server handle the default visibility by sending an empty string if flag not set
	repoCreateCmd.Flags().StringVar(&createVisibilityFlag, "visibility", "", "Repository visibility: public, internal, private (defaults to organization/server setting)")
}
