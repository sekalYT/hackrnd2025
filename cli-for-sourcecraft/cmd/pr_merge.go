// cmd/pr_merge.go
package cmd

import (
	"fmt"
	"strings"

	"cli-for-sourcecraft/internal/git"
	cliutils "cli-for-sourcecraft/internal/utils"

	"cli-for-sourcecraft/internal/api" // Нужно для MergeParameters

	"github.com/spf13/cobra"
)

// Flags for pr merge
var (
	prMergeRepoFlag         string
	prMergeSquashFlag       bool // --squash
	prMergeRebaseFlag       bool // --rebase
	prMergeDeleteBranchFlag bool // --delete-branch
)

var prMergeCmd = &cobra.Command{
	Use:   "merge <pr_id_or_slug>",
	Short: "Merge a pull request into its target branch",
	Long: `Merges a pull request. This requires the PR to be approved and ready for merge.

Example: src pr merge 1 --squash --delete-branch`,
	Args: cobra.ExactArgs(1), // Требуем <pr_id_or_slug>
	RunE: func(cmd *cobra.Command, args []string) error {
		prSlug := args[0]
		var orgSlug, repoSlug string
		var err error

		// 1. Определяем Репозиторий
		if prMergeRepoFlag != "" {
			parts := strings.SplitN(prMergeRepoFlag, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid format for --repo flag: '%s'. Expected format: <org_slug>/<repo_slug>", prMergeRepoFlag)
			}
			orgSlug = parts[0]
			repoSlug = parts[1]
		} else {
			// Пытаемся получить из текущей директории git
			orgSlug, repoSlug, err = git.GetCurrentRepoOwnerAndNameFromRemote("origin")
			if err != nil {
				return fmt.Errorf("could not detect repository from git remote. Use --repo <org>/<repo> flag or run from within a repository")
			}
		}

		// *** ИЗМЕНЕНИЕ: ***
		// Проверка на конфликты стратегий ИЛИ ЛЮБОЙ ИЗ ФЛАГОВ
		if prMergeSquashFlag && prMergeRebaseFlag {
			return fmt.Errorf("cannot use --squash and --rebase together. Choose one.")
		}

		// *** ДОБАВЛЕНО: Предупреждение, если API не поддерживает флаги ***
		if prMergeSquashFlag || prMergeRebaseFlag || prMergeDeleteBranchFlag {
			fmt.Println("---------------------------------------------------------------------")
			fmt.Println("ВНИМАНИЕ: Флаги --squash, --rebase и --delete-branch")
			fmt.Println("не поддерживаются текущей спецификацией API (Swagger).")
			fmt.Println("Команда попытается 'одобрить' (approve) PR, но")
			fmt.Println("параметры слияния будут проигнорированы сервером.")
			fmt.Println("---------------------------------------------------------------------")
		}

		// 2. Готовим параметры слияния
		// *ПРИМЕЧАНИЕ*: Мы все еще передаем их в apiClient,
		// но (согласно нашему исправлению в client.go) они будут проигнорированы.
		mergeParams := api.MergeParameters{
			Squash:       prMergeSquashFlag,
			Rebase:       prMergeRebaseFlag,
			DeleteBranch: prMergeDeleteBranchFlag,
		}

		// 3. Вызываем API
		// *** ИЗМЕНЕНИЕ: Текст сообщения, так как мы 'одобряем', а не 'мержим' ***
		fmt.Printf("Attempting to 'approve' (merge) PR #%s in %s/%s...\n", prSlug, orgSlug, repoSlug)

		// *** ИЗМЕНЕНИЕ: ***
		// apiClient.MergePullRequest ТЕПЕРЬ ВОЗВРАЩАЕТ (*SetDecisionResponse, error)
		//
		decisionResponse, err := apiClient.MergePullRequest(orgSlug, repoSlug, prSlug, mergeParams)
		if err != nil {
			return err // Ошибка API (например, 404, если PR не найден)
		}

		// 4. Выводим результат
		// *** ИЗМЕНЕНИЕ: Мы больше не получаем 'mergedPR.Status' ***
		// Мы получаем 'decisionResponse.CreatedDecision'

		newDecision := cliutils.DerefString(decisionResponse.CreatedDecision)

		fmt.Println("\nAPI request complete.")
		fmt.Printf("PR #%s decision set to: %s\n", prSlug, newDecision)

		if newDecision == "approve" {
			fmt.Println("SUCCESS: Pull Request 'approve' decision was set.")
			fmt.Println("Сервер должен автоматически запустить слияние, если все проверки пройдены.")
		} else {
			fmt.Printf("INFO: API returned decision '%s'.\n", newDecision)
		}

		return nil
	},
}

func init() {
	prCmd.AddCommand(prMergeCmd) // Добавляем 'merge' к 'pr'

	// Добавляем флаги
	prMergeCmd.Flags().StringVarP(&prMergeRepoFlag, "repo", "R", "", "Specify repository in <org>/<repo> format (default: current directory)")
	prMergeCmd.Flags().BoolVar(&prMergeSquashFlag, "squash", false, "Use squash merge strategy (NOTE: Not supported by current API)")
	prMergeCmd.Flags().BoolVar(&prMergeRebaseFlag, "rebase", false, "Use rebase merge strategy (NOTE: Not supported by current API)")
	prMergeCmd.Flags().BoolVar(&prMergeDeleteBranchFlag, "delete-branch", false, "Delete the source branch after a successful merge (NOTE: Not supported by current API)")
}
