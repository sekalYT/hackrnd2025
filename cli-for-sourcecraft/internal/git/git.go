// internal/git/git.go
package git

import (
	"cli-for-sourcecraft/internal/api"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// GetCurrentRepoOwnerAndNameFromRemote парсит URL указанного remote
func GetCurrentRepoOwnerAndNameFromRemote(remoteName string) (owner string, repo string, err error) {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	output, err := cmd.Output()
	if err != nil {
		// Стандартный вывод git для 'no such remote' может идти в stderr или stdout
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		if strings.Contains(strings.ToLower(string(output)), "no such remote") || strings.Contains(strings.ToLower(stderr), "no such remote") {
			return "", "", fmt.Errorf("git remote '%s' not found", remoteName)
		}
		return "", "", fmt.Errorf("failed to get URL for remote '%s': %w. Output: %s, Stderr: %s", remoteName, err, string(output), stderr)
	}

	remoteURL := strings.TrimSpace(string(output))
	return ParseOwnerAndRepoFromURL(remoteURL) // Используем общую функцию парсинга
}

// GetRemotes возвращает map существующих remotes и их URL
func GetRemotes() (map[string]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run 'git remote -v': %w", err)
	}

	remotes := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		// Ожидаем строки вида: origin <url> (fetch/push)
		if len(fields) >= 3 && (fields[2] == "(fetch)" || fields[2] == "(push)") { // Более точная проверка
			remoteName := fields[0]
			remoteURL := fields[1]
			// Сохраняем fetch URL (обычно он первый)
			if _, exists := remotes[remoteName]; !exists && strings.Contains(line, "(fetch)") {
				remotes[remoteName] = remoteURL
			} else if _, exists := remotes[remoteName]; !exists { // Или любой, если fetch нет
				remotes[remoteName] = remoteURL
			}
		}
	}
	return remotes, nil
}

// GetRemoteURL возвращает URL для указанного remote
func GetRemoteURL(remoteName string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	output, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		// Улучшенная проверка на 'no such remote'
		if strings.Contains(strings.ToLower(string(output)), "no such remote") || strings.Contains(strings.ToLower(stderr), "no such remote") {
			return "", fmt.Errorf("remote '%s' not found", remoteName)
		}
		return "", fmt.Errorf("failed to get URL for remote '%s': %w. Output: %s, Stderr: %s", remoteName, err, string(output), stderr)
	}
	return strings.TrimSpace(string(output)), nil
}

// ParseOwnerAndRepoFromURL пытается извлечь owner/repo из SSH или HTTPS URL
func ParseOwnerAndRepoFromURL(remoteURL string) (owner string, repo string, err error) {
	remoteURL = strings.TrimSpace(remoteURL)
	var repoPath string

	if strings.HasPrefix(remoteURL, "https://") {
		// Убираем возможные credentials из URL перед парсингом
		cleanURL := remoteURL
		if idx := strings.Index(remoteURL, "@"); idx > strings.Index(remoteURL, "://")+3 {
			protoEnd := strings.Index(remoteURL, "://") + 3
			cleanURL = remoteURL[:protoEnd] + remoteURL[idx+1:]
		}
		u, err := url.Parse(cleanURL) // Парсим очищенный URL
		if err != nil {
			return "", "", fmt.Errorf("could not parse HTTPS URL '%s': %w", remoteURL, err)
		}
		repoPath = u.Path
	} else if strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://") {
		var hostAndPath string
		if strings.HasPrefix(remoteURL, "ssh://") {
			// ssh://user@host:port/path.git -> /path.git
			u, err := url.Parse(remoteURL)
			if err != nil {
				return "", "", fmt.Errorf("could not parse ssh:// URL '%s': %w", remoteURL, err)
			}
			repoPath = u.Path
			// parts := strings.SplitN(remoteURL, "/", 4) // ssh:, , user@host, path
			// if len(parts) < 4 {
			// 	return "", "", fmt.Errorf("could not parse ssh:// URL '%s'", remoteURL)
			// }
			// hostAndPath = parts[3] // path/to/repo.git
		} else {
			// git@host:path.git -> path.git
			parts := strings.SplitN(remoteURL, ":", 2)
			if len(parts) != 2 {
				return "", "", fmt.Errorf("could not parse git@ URL '%s'", remoteURL)
			}
			hostAndPath = parts[1]       // organization-sekal01/my-repo.git
			repoPath = "/" + hostAndPath // Добавляем слэш для единообразия
		}
		// repoPath = "/" + hostAndPath // Add slash for consistency

	} else {
		return "", "", fmt.Errorf("unsupported remote URL format '%s'", remoteURL)
	}

	// Извлекаем owner/repo из пути
	// /organization-sekal01/my-repo.git -> organization-sekal01/my-repo
	fullSlug := strings.TrimPrefix(repoPath, "/")
	fullSlug = strings.TrimSuffix(fullSlug, ".git")

	parts := strings.SplitN(fullSlug, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("could not extract owner/repo from path '%s' (parsed from URL '%s')", repoPath, remoteURL)
	}

	return parts[0], parts[1], nil
}

func GetCurrentBranchName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return "", fmt.Errorf("failed to get current branch name: %w. Output: %s, Stderr: %s", err, string(output), stderr)
	}
	branchName := strings.TrimSpace(string(output))
	if branchName == "HEAD" {
		return "", fmt.Errorf("currently in detached HEAD state, not on a branch")
	}
	return branchName, nil
}

func GetDefaultBranchName(apiRepoInfo *api.Repo, remoteName string) (string, error) {
	// 1. Из API (если есть)
	if apiRepoInfo != nil && apiRepoInfo.DefaultBranch != nil && *apiRepoInfo.DefaultBranch != "" {
		return *apiRepoInfo.DefaultBranch, nil
	}

	// 2. Из git remote show
	cmd := exec.Command("git", "remote", "show", remoteName)
	output, err := cmd.Output()
	if err != nil {
		// Если repoInfo не было и remote не найден - не страшно, вернем дефолт
		if apiRepoInfo == nil && strings.Contains(err.Error(), "not found") {
			fmt.Println("Warning: Could not determine default branch via git remote show (remote not found?). Assuming 'main'.")
			return "main", nil // Возвращаем дефолт, а не ошибку
		}
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		// Возвращаем ошибку, если remote должен был существовать
		return "", fmt.Errorf("failed to get remote info for '%s' to determine default branch: %w. Output: %s, Stderr: %s", remoteName, err, string(output), stderr)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "HEAD branch:") {
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) == 2 {
				branch := strings.TrimSpace(parts[1])
				if branch != "" && branch != "(unknown)" {
					return branch, nil
				}
			}
		}
	}

	// 3. Финальный дефолт
	fmt.Println("Warning: Could not determine default branch via git remote show. Assuming 'main'.")
	return "main", nil
}

// GetLastCommitTitle получает заголовок последнего коммита в ветке/ref.
func GetLastCommitTitle(refName string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%s", refName)
	output, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return "", fmt.Errorf("failed to get last commit title for '%s': %w. Stderr: %s", refName, err, stderr)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCommitMessagesSinceBase получает заголовки коммитов с момента расхождения веток.
func GetCommitMessagesSinceBase(baseBranch, headBranch string) (string, error) {
	rangeSpec := fmt.Sprintf("%s..%s", baseBranch, headBranch)
	cmd := exec.Command("git", "log", "--pretty=%s", rangeSpec) // Только заголовки (%s)
	// Для полного сообщения: --pretty=%B
	output, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		// Не фатально, можем просто не предлагать тело PR
		fmt.Printf("Warning: failed to get commit messages for range '%s': %v. Stderr: %s\n", rangeSpec, err, stderr)
		return "", nil // Возвращаем пустую строку, а не ошибку
	}
	return strings.TrimSpace(string(output)), nil // Возвращаем все заголовки
}
