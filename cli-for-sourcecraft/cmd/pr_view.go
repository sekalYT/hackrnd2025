// cmd/pr_view.go
package cmd

import (
	"fmt"
	"strings"
	"time"

	"cli-for-sourcecraft/internal/api"
	"cli-for-sourcecraft/internal/git"            // Import git for getting current repo
	cliutils "cli-for-sourcecraft/internal/utils" // Import utils

	"github.com/spf13/cobra"
)

// Flags for pr view
var (
	prViewRepoFlag string // Flag to specify repo, e.g., --repo my-org/my-repo
)

var prViewCmd = &cobra.Command{
	Use:   "view <pr_id_or_slug>",
	Short: "View detailed information about a pull request",
	Long: `Displays detailed information about a pull request (PR).

The <pr_id_or_slug> is typically the PR number (slug) or its internal ID.
If the repository is not specified with --repo, it uses the current git repository.`,
	Args: cobra.ExactArgs(1), // Требуем <pr_id_or_slug>
	RunE: func(cmd *cobra.Command, args []string) error {
		prSlug := args[0] // ID или номер PR
		var orgSlug, repoSlug string
		var err error

		// 1. Определяем Репозиторий
		if prViewRepoFlag != "" {
			// Используем репозиторий из флага
			parts := strings.SplitN(prViewRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid format for --repo flag: '%s'. Expected format: <org_slug>/<repo_slug>", prViewRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
			fmt.Printf("Viewing PR %s in specified repository: %s/%s\n", prSlug, orgSlug, repoSlug)
		} else {
			// Пытаемся получить из текущей директории git
			fmt.Println("Attempting to detect repository from git remote 'origin'...")
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("could not detect repository from git remote. Use --repo <org>/<repo> flag or run from within a repository")
			}
			fmt.Printf("Viewing PR %s in detected repository: %s/%s\n", prSlug, orgSlug, repoSlug)
		}

		// 2. Вызываем API для получения PR
		pr, err := apiClient.GetPullRequest(orgSlug, repoSlug, prSlug)
		if err != nil {
			return err // Ошибка 404 или другая будет выведена
		}

		// 3. Выводим информацию красиво
		fmt.Println("--- Pull Request Details ---")
		fmt.Printf("Title:       %s\n", cliutils.DerefString(pr.Title))
		fmt.Printf("ID/Slug:     %s\n", cliutils.DerefString(pr.Slug))
		fmt.Printf("Status:      %s\n", cliutils.DerefString(pr.Status))

		// Информация о ветках
		fmt.Printf("Source Br:   %s\n", cliutils.DerefString(pr.SourceBranch))
		fmt.Printf("Target Br:   %s\n", cliutils.DerefString(pr.TargetBranch))
		fmt.Printf("Author:      %s\n", getAuthor(pr))

		// Форматирование даты
		if pr.UpdatedAt != nil {
			updatedAtStr := cliutils.DerefString(pr.UpdatedAt)
			t, parseErr := time.Parse(time.RFC3339Nano, updatedAtStr)
			if parseErr != nil {
				t, parseErr = time.Parse(time.RFC3339, updatedAtStr)
			}
			if parseErr == nil {
				fmt.Printf("Last Update: %s (%s ago)\n", t.Local().Format("2006-01-02 15:04:05 MST"), formatTimeAgo(t))
			} else {
				fmt.Printf("Last Update: %s (raw)\n", updatedAtStr)
			}
		}

		// Формируем URL
		webURL := fmt.Sprintf("https://sourcecraft.dev/%s/%s/pr/%s", orgSlug, repoSlug, prSlug)
		fmt.Printf("View online: %s\n", webURL)

		// Выводим описание (если есть)
		// NOTE: Нам нужно, чтобы PullRequest struct содержала поле Description
		// Так как мы не знаем, возвращает ли GetPullRequest Description,
		// попробуем вывести его, если оно есть в структуре.
		// fmt.Println("\n--- Description ---")
		// fmt.Println(cliutils.DerefString(pr.Description))

		return nil
	},
}

// getAuthor - хелпер для безопасного получения slug автора
func getAuthor(pr *api.PullRequest) string {
	if pr.Author != nil {
		return cliutils.DerefString(pr.Author.Slug)
	}
	return "-"
}

// formatTimeAgo - хелпер для форматирования времени (дубликат formatRelativeTime)
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration.Minutes() < 60 {
		return fmt.Sprintf("%d minutes", int(duration.Minutes()))
	}
	if duration.Hours() < 24 {
		return fmt.Sprintf("%d hours", int(duration.Hours()))
	}
	return fmt.Sprintf("%d days", int(duration.Hours()/24))
}

func init() {
	prCmd.AddCommand(prViewCmd) // Добавляем 'view' к 'pr'
	// Добавляем флаг --repo
	prViewCmd.Flags().StringVarP(&prViewRepoFlag, "repo", "R", "", "Specify repository in <org>/<repo> format (default: current directory)")

	// Убедимся, что formatRelativeTime удален из pr_list.go, чтобы не было конфликтов
	// и используем formatTimeAgo или formatRelativeTime из utils, если его туда вынесли.
}
