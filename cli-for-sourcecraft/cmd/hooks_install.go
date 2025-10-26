// cmd/hooks_install.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var supportedHooks = []string{"pre-commit", "pre-push"}

var hooksInstallCmd = &cobra.Command{
	Use:   "install <hook-type>",
	Short: "Install a Git hook script",
	Long:  "Installs a script for the specified hook type (e.g., pre-commit) into .git/hooks.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hookType := args[0]

		isValidHook := false
		for _, supported := range supportedHooks {
			if hookType == supported {
				isValidHook = true
				break
			}
		}
		if !isValidHook {
			return fmt.Errorf("unsupported hook type '%s'. Supported types are: %s", hookType, strings.Join(supportedHooks, ", "))
		}

		gitDir, err := findGitDir()
		if err != nil {
			return fmt.Errorf("current directory is not inside a Git repository")
		}
		hooksDir := filepath.Join(gitDir, "hooks")
		hookFilePath := filepath.Join(hooksDir, hookType)

		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return fmt.Errorf("failed to create hooks directory '%s': %w", hooksDir, err)
		}

		var scriptContent string
		switch hookType {
		case "pre-commit":
			scriptContent = `#!/bin/sh
# Hook installed by src CLI

echo "Running src pre-commit checks..."
# Add your src command here
# src security lint --staged # Example

exit 0 # Allow commit
`
		case "pre-push":
			scriptContent = `#!/bin/sh
# Hook installed by src CLI

remote="$1"
url="$2"

echo "Running src pre-push checks for remote $remote ($url)..."
# Add your src command here
# src pr check-status $(git rev-parse --abbrev-ref HEAD) # Example

exit 0 # Allow push
`
		default:
			return fmt.Errorf("internal error: script content not defined for %s", hookType)
		}

		fmt.Printf("Writing hook script to: %s\n", hookFilePath)
		err = os.WriteFile(hookFilePath, []byte(scriptContent), 0755)
		if err != nil {
			return fmt.Errorf("failed to write hook script '%s': %w", hookFilePath, err)
		}

		if runtime.GOOS != "windows" {
			err = os.Chmod(hookFilePath, 0755)
			if err != nil {
				fmt.Printf("Warning: failed to make hook script executable '%s': %v\n", hookFilePath, err)
			}
		}

		fmt.Printf("Successfully installed '%s' hook.\n", hookType)
		return nil
	},
}

func findGitDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		gitDirPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitDirPath)
		if err == nil && info.IsDir() {
			return gitDirPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(".git directory not found")
}

func init() {
	hooksCmd.AddCommand(hooksInstallCmd)
}
