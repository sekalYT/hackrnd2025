// cmd/repo_view.go
package cmd

import (
	"fmt"
	"time" // Для форматирования даты

	cliutils "cli-for-sourcecraft/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "cli-for-sourcecraft/internal/api" // Не нужен прямой импорт api
)

var repoViewCmd = &cobra.Command{
	Use:   "view <repo_slug>",
	Short: "View information about a specific repository",
	Long: `Displays detailed information about a repository within the organization specified in config.yaml.
Example: src repo view my-awesome-repo`,
	Args: cobra.ExactArgs(1), // Требуем <repo_slug>
	RunE: func(cmd *cobra.Command, args []string) error {
		repoSlug := args[0]

		orgSlug := viper.GetString("organization")
		if orgSlug == "" {
			return fmt.Errorf("error: 'organization' slug not found in config.yaml or SOURCECRAFT_ORGANIZATION env var")
		}

		fmt.Printf("Fetching details for repository '%s/%s'...\n", orgSlug, repoSlug)

		// apiClient уже создан в rootCmd
		repo, err := apiClient.GetRepository(orgSlug, repoSlug)
		if err != nil {
			return err // Ошибка 404 или другая будет выведена
		}

		// Выводим информацию красиво
		fmt.Println("--- Repository Details ---")
		fmt.Printf("Name:        %s\n", cliutils.DerefString(repo.Name))
		fmt.Printf("Slug:        %s\n", cliutils.DerefString(repo.Slug))
		fmt.Printf("Full Path:   %s/%s\n", orgSlug, cliutils.DerefString(repo.Slug))
		fmt.Printf("Visibility:  %s\n", cliutils.DerefString(repo.Visibility))
		fmt.Printf("Description: %s\n", cliutils.DerefString(repo.Description))
		fmt.Printf("Default Br:  %s\n", cliutils.DerefString(repo.DefaultBranch))
		fmt.Printf("Is Empty:    %t\n", cliutils.DerefBool(repo.IsEmpty))

		// Выводим URL для клонирования
		if repo.CloneURL != nil {
			fmt.Printf("SSH URL:     %s\n", cliutils.DerefString(repo.CloneURL.SSH))
			fmt.Printf("HTTPS URL:   %s\n", cliutils.DerefString(repo.CloneURL.HTTPS))
		}

		// Выводим информацию о владельце (если есть)
		if repo.Owner != nil {
			//
			// *** ИСПРАВЛЕНИЕ ЗДЕСЬ: Используем Slug вместо Username ***
			//
			fmt.Printf("Owner Slug:  %s\n", cliutils.DerefString(repo.Owner.Slug)) // Используем Slug, как в client.go и Swagger
		}

		// Форматируем дату последнего обновления
		if repo.LastUpdated != nil {
			// Swagger не указывает формат Timestamp, пробуем стандартный RFC3339
			t, err := time.Parse(time.RFC3339Nano, *repo.LastUpdated) // Пробуем RFC3339Nano для большей точности
			if err != nil {
				// Если Nano не сработал, пробуем обычный RFC3339
				t, err = time.Parse(time.RFC3339, *repo.LastUpdated)
			}

			if err == nil {
				fmt.Printf("Last Update: %s\n", t.Local().Format("2006-01-02 15:04:05 MST"))
			} else {
				fmt.Printf("Last Update: %s (raw, parse error: %v)\n", *repo.LastUpdated, err) // Показываем ошибку парсинга
			}
		}

		// Выводим язык (если есть)
		if repo.Language != nil {
			fmt.Printf("Language:    %s\n", cliutils.DerefString(repo.Language.Name))
		}

		return nil
	},
}

// cliutils.DerefString уже есть в других файлах, вынести в utils!
// func cliutils.DerefString(s *string) string {
// 	if s != nil { return *s }
// 	return ""
// }

func init() {
	repoCmd.AddCommand(repoViewCmd) // Добавляем 'view' как подкоманду 'repo'
}
