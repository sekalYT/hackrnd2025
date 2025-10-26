// cmd/auth_login.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authentication using a personal access token (PAT)",
	Long: `Requests your Personal Access Token
and stores it in a configuration file for subsequent API requests.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your Personal Access Token (PAT): ")
		token, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("the token could not be read: %w", err)
		}
		token = strings.TrimSpace(token)

		if token == "" {
			return fmt.Errorf("the token cannot be empty")
		}

		err = keyring.Set(keyringServiceName, keyringTokenUser, token)
		if err != nil {
			return fmt.Errorf("failed to save the token in OS keyring: %w", err)
		}

		if viper.IsSet("token") {
			viper.Set("token", nil)
		}

		err = viper.WriteConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				configDir := getConfigDir()
				configPath := filepath.Join(configDir, "config.yaml")
				fmt.Printf("The configuration file is not found, so we create a new one: %s\n", configPath)
				if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create configuration directory '%s': %w", configDir, err)
				}
				err = viper.SafeWriteConfig()
				if err != nil {
					err = viper.WriteConfigAs(configPath)
					if err != nil {
						return fmt.Errorf("failed to create and save configuration file '%s': %w", configPath, err)
					}
				}
			} else {
				return fmt.Errorf("could not save the configuration: %w", err)
			}
		}

		fmt.Println("\nAuthentication was successful. The token is stored in a secure vault (OS keyring).")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
