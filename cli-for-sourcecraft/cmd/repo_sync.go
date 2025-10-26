// cmd/repo_sync.go
package cmd

import (
	"bufio" // <-- Для чтения ввода пользователя
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cli-for-sourcecraft/internal/api" // <-- Нужен только для api.Repo при запросе upstream URL
	"cli-for-sourcecraft/internal/git" // Импортируем наш git хелпер

	"github.com/spf13/cobra"
	// "github.com/spf13/viper" // Не нужен здесь
)

// Флаги для команды sync
var (
	syncPushFlag  bool
	syncHTTPSFlag bool
)

var repoSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a forked repository with its upstream parent",
	Long: `Fetches changes from the original (upstream) repository and merges them
into the current branch (usually main or master) of your local fork.

If the 'upstream' remote is not configured, you will be prompted to enter the
full path of the original repository (e.g., original-owner/original-repo).
The --https flag forces using the HTTPS URL for the upstream remote.
This command must be run from within the root directory of your cloned fork.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		// 1. Определяем текущий репозиторий (наш форк) - опционально, для default branch
		fmt.Println("Checking current repository...")
		currentOrg, currentRepo, err := git.GetCurrentRepoOwnerAndNameFromRemote("origin")
		if err != nil {
			// Не фатально, можем не знать org/repo форка
			fmt.Printf("Warning: could not determine current repository from 'origin' remote: %v. Will assume default branch is 'main'.\n", err)
			currentOrg = "" // Сбрасываем, чтобы не использовать дальше
			currentRepo = ""
		} else {
			fmt.Printf("Detected current repository: %s/%s\n", currentOrg, currentRepo)
		}

		// 2. Получаем дефолтную ветку текущего репо (если смогли определить его)
		defaultBranch := "main" // Запасной вариант
		if currentOrg != "" && currentRepo != "" {
			fmt.Println("Fetching repository details to determine default branch...")
			repoInfo, err := apiClient.GetRepository(currentOrg, currentRepo)
			if err != nil {
				fmt.Printf("Warning: failed to get repository info for %s/%s: %v. Assuming default branch is 'main'.\n", currentOrg, currentRepo, err)
			} else if repoInfo != nil && repoInfo.DefaultBranch != nil && *repoInfo.DefaultBranch != "" {
				defaultBranch = *repoInfo.DefaultBranch
			}
		}
		fmt.Printf("Target local branch for merge: '%s'\n", defaultBranch)

		// 3. Проверяем/Настраиваем 'upstream' remote
		desiredUpstreamURL, err := ensureUpstreamRemote(syncHTTPSFlag) // Выносим логику в хелпер
		if err != nil {
			return err
		}
		fmt.Printf("Using upstream URL: %s\n", desiredUpstreamURL)

		// 4. Fetch changes from upstream
		fmt.Println("Fetching changes from upstream...")
		if err := runGitCommand("fetch", "upstream"); err != nil {
			return fmt.Errorf("failed to fetch from upstream: %w", err)
		}

		// 5. Checkout local default branch
		fmt.Printf("Switching to local branch '%s'...\n", defaultBranch)
		if err := runGitCommand("checkout", defaultBranch); err != nil {
			return fmt.Errorf("failed to checkout branch '%s': %w. Make sure you don't have uncommitted changes", defaultBranch, err)
		}

		// 6. Merge upstream changes
		// Предполагаем, что у upstream дефолтная ветка называется так же.
		upstreamDefaultBranchRef := fmt.Sprintf("upstream/%s", defaultBranch)
		fmt.Printf("Merging changes from '%s' into '%s'...\n", upstreamDefaultBranchRef, defaultBranch)
		err = runGitCommand("merge", "--no-ff", upstreamDefaultBranchRef) // Используем --no-ff для явного merge коммита
		// Проверяем ошибки мержа (включая конфликты)
		if err != nil {
			// Проверяем статус на наличие конфликтов
			conflictOutput, statusErr := exec.Command("git", "status", "--porcelain").Output()
			isConflict := statusErr == nil && strings.Contains(string(conflictOutput), "UU ")

			if isConflict {
				fmt.Println("\n---")
				fmt.Println("Warning: Merge resulted in conflicts!")
				fmt.Println("Please resolve the conflicts manually:")
				fmt.Println("  1. Edit the conflicted files (look for '<<<<<<<', '=======', '>>>>>>>').")
				fmt.Println("  2. Run 'git add <resolved-files>' for each resolved file.")
				fmt.Println("  3. Run 'git commit' to finalize the merge.")
				fmt.Println("After resolving, you can optionally push to your origin.")
				fmt.Println("---")
				// Возвращаем nil, т.к. команда инициировала процесс, но требует ручного вмешательства
				return nil
			} else {
				// Другая ошибка мержа
				return fmt.Errorf("merge from '%s' failed: %w", upstreamDefaultBranchRef, err)
			}
		}
		fmt.Println("Merge successful or already up-to-date.")

		// 7. Push changes (optional)
		if syncPushFlag {
			fmt.Printf("Pushing updated branch '%s' to origin...\n", defaultBranch)
			if err := runGitCommand("push", "origin", defaultBranch); err != nil {
				return fmt.Errorf("failed to push to origin: %w", err)
			}
			fmt.Println("Push successful.")
		} else if err == nil { // Предлагаем push только если мерж прошел чисто
			fmt.Printf("\nSync complete locally. Run 'git push origin %s' to update your remote fork on SourceCraft.\n", defaultBranch)
		}

		return nil
	},
}

// ensureUpstreamRemote проверяет remote 'upstream', при необходимости запрашивает путь,
// получает нужный URL (SSH/HTTPS) и настраивает/обновляет remote.
func ensureUpstreamRemote(useHTTPS bool) (string, error) {
	existingUpstreamURL, err := git.GetRemoteURL("upstream")
	upstreamNotFound := (err != nil && strings.Contains(err.Error(), "not found"))

	if err != nil && !upstreamNotFound {
		return "", fmt.Errorf("failed to check 'upstream' remote: %w", err)
	}

	desiredProto := "SSH"
	if useHTTPS {
		desiredProto = "HTTPS"
	}

	if !upstreamNotFound {
		fmt.Printf("Existing 'upstream' remote found: %s\n", existingUpstreamURL)
		// Проверяем, совпадает ли протокол
		isExistingSSH := strings.HasPrefix(existingUpstreamURL, "git@") || strings.HasPrefix(existingUpstreamURL, "ssh://")
		needsUpdate := (useHTTPS && isExistingSSH) || (!useHTTPS && !isExistingSSH)

		if needsUpdate {
			fmt.Printf("Protocol mismatch detected (Existing is %s, requested %s). Updating URL...\n", map[bool]string{true: "SSH", false: "HTTPS"}[isExistingSSH], desiredProto)
			// Парсим owner/repo из существующего URL, чтобы запросить новый
			upstreamOrgSlug, upstreamRepoSlug, parseErr := git.ParseOwnerAndRepoFromURL(existingUpstreamURL)
			if parseErr != nil {
				// Если не смогли распарсить, просим пользователя ввести заново
				fmt.Printf("Warning: Could not parse existing upstream URL to update protocol: %v\n", parseErr)
				upstreamOrgSlug, upstreamRepoSlug, parseErr = promptForUpstreamPath(nil) // Передаем nil, т.к. repoInfo опционален
				if parseErr != nil {
					return "", parseErr
				}
			}
			// Запрашиваем URL нужного протокола
			newUpstreamURL, fetchErr := fetchUpstreamURL(upstreamOrgSlug, upstreamRepoSlug, useHTTPS)
			if fetchErr != nil {
				return "", fetchErr
			}

			// Обновляем remote
			fmt.Printf("Updating 'upstream' remote URL to %s...\n", newUpstreamURL)
			if err := runGitCommand("remote", "set-url", "upstream", newUpstreamURL); err != nil {
				return "", fmt.Errorf("failed to update 'upstream' remote URL: %w", err)
			}
			return newUpstreamURL, nil // Возвращаем обновленный URL
		} else {
			fmt.Println("Existing upstream URL protocol matches.")
			return existingUpstreamURL, nil // Используем существующий URL
		}

	} else {
		// Upstream не найден, нужно добавить
		fmt.Println("'upstream' remote not found.")
		// Запрашиваем путь у пользователя
		upstreamOrgSlug, upstreamRepoSlug, promptErr := promptForUpstreamPath(nil) // Передаем nil, т.к. repoInfo опционален здесь
		if promptErr != nil {
			return "", promptErr
		}

		// Запрашиваем URL нужного протокола
		newUpstreamURL, fetchErr := fetchUpstreamURL(upstreamOrgSlug, upstreamRepoSlug, useHTTPS)
		if fetchErr != nil {
			return "", fetchErr
		}

		// Добавляем remote
		fmt.Printf("Adding 'upstream' remote pointing to %s...\n", newUpstreamURL)
		if err := runGitCommand("remote", "add", "upstream", newUpstreamURL); err != nil {
			return "", fmt.Errorf("failed to add 'upstream' remote: %w", err)
		}
		return newUpstreamURL, nil // Возвращаем добавленный URL
	}
}

// Helper function to prompt user for upstream path (repoInfo can be nil)
func promptForUpstreamPath(repoInfo *api.Repo) (orgSlug, repoSlug string, err error) {
	parentSlugHint := "original-repo" // Generic hint
	if repoInfo != nil && repoInfo.Parent != nil && repoInfo.Parent.Slug != nil && *repoInfo.Parent.Slug != "" {
		parentSlugHint = *repoInfo.Parent.Slug
	}

	fmt.Printf("Please enter the full path of the original repository you forked from\n(e.g., original-owner/%s): ", parentSlugHint)
	reader := bufio.NewReader(os.Stdin)
	upstreamPathInput, _ := reader.ReadString('\n')
	upstreamPathInput = strings.TrimSpace(upstreamPathInput)

	parts := strings.SplitN(upstreamPathInput, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid format. Please enter in the format 'owner/repo'")
	}
	return parts[0], parts[1], nil
}

// Helper function to fetch the correct upstream URL from API
func fetchUpstreamURL(orgSlug, repoSlug string, useHTTPS bool) (string, error) {
	fmt.Printf("Fetching details for upstream repository %s/%s to get clone URL...\n", orgSlug, repoSlug)
	upstreamRepoInfo, err := apiClient.GetRepository(orgSlug, repoSlug) // Используем GetRepository
	if err != nil {
		return "", fmt.Errorf("failed to get upstream repository info (%s/%s) for clone URL: %w", orgSlug, repoSlug, err)
	}
	if upstreamRepoInfo.CloneURL == nil {
		return "", fmt.Errorf("API did not provide clone URLs for upstream repository %s/%s", orgSlug, repoSlug)
	}

	var url string
	if useHTTPS {
		if upstreamRepoInfo.CloneURL.HTTPS != nil && *upstreamRepoInfo.CloneURL.HTTPS != "" {
			url = *upstreamRepoInfo.CloneURL.HTTPS
			fmt.Println("Using HTTPS URL for upstream.")
		} else if upstreamRepoInfo.CloneURL.SSH != nil && *upstreamRepoInfo.CloneURL.SSH != "" { // Fallback
			url = *upstreamRepoInfo.CloneURL.SSH
			fmt.Println("HTTPS URL not found for upstream, falling back to SSH.")
		}
	} else { // Use SSH by default
		if upstreamRepoInfo.CloneURL.SSH != nil && *upstreamRepoInfo.CloneURL.SSH != "" {
			url = *upstreamRepoInfo.CloneURL.SSH
			fmt.Println("Using SSH URL for upstream.")
		} else if upstreamRepoInfo.CloneURL.HTTPS != nil && *upstreamRepoInfo.CloneURL.HTTPS != "" { // Fallback
			url = *upstreamRepoInfo.CloneURL.HTTPS
			fmt.Println("SSH URL not found for upstream, falling back to HTTPS.")
		}
	}

	if url == "" {
		return "", fmt.Errorf("no suitable clone URL (SSH or HTTPS) available for upstream %s/%s", orgSlug, repoSlug)
	}
	return url, nil
}

// runGitCommand helper
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	// Перенаправляем stdout/stderr в os.Stdout/os.Stderr, чтобы пользователь видел вывод git
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Running: git %s\n", strings.Join(args, " ")) // Логируем команду перед запуском
	err := cmd.Run()
	// Не оборачиваем ошибку, чтобы сохранить ExitError
	return err
}

func init() {
	repoCmd.AddCommand(repoSyncCmd)
	repoSyncCmd.Flags().BoolVar(&syncPushFlag, "push", false, "Push the updated branch to your fork ('origin') after merging")
	repoSyncCmd.Flags().BoolVar(&syncHTTPSFlag, "https", false, "Use HTTPS URL for the upstream remote instead of SSH (used when adding or updating upstream)")
}

// derefString (нужен)
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// derefBool (нужен)
func derefBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}
