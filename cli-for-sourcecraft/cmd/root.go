// cmd/root.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cli-for-sourcecraft/internal/api"

	"github.com/zalando/go-keyring"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiClient *api.Client
var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "src",
	Short: "CLI для SourceCraft.dev",
	Long: `A command-line tool (CLI) for the SourceCraft.dev platform.
Provides access to repositories, pull requests, tasks, and more.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		if cmd.Name() == "help" {
			return nil
		}
		if cmd.Parent() != nil && cmd.Parent().Name() == "auth" {
			return nil
		}

		var token string
		var err error

		token = os.Getenv("SOURCECRAFT_TOKEN")
		if token != "" {
			apiClient = api.NewClient("https://api.sourcecraft.tech", token)
			return nil
		}

		token, err = keyring.Get(keyringServiceName, keyringTokenUser)
		if err != nil && err != keyring.ErrNotFound {
			return fmt.Errorf("error reading OS keyring: %w", err)
		}

		if err == keyring.ErrNotFound {
			yamlToken := viper.GetString("token")

			if yamlToken != "" {
				fmt.Fprintln(os.Stderr, "Token detected in config.yaml, migration to secure storage (OS keyring)...")

				errSet := keyring.Set(keyringServiceName, keyringTokenUser, yamlToken)
				if errSet != nil {
					return fmt.Errorf("failed to migrate token to keyring: %w", errSet)
				}

				viper.Set("token", nil)
				if errWrite := viper.WriteConfig(); errWrite != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to remove token from config.yaml after migration: %v\n", errWrite)
				}

				token = yamlToken
				fmt.Fprintln(os.Stderr, "The migration is complete. The token is now in OS keyring.")
			}
		}

		if token == "" {
			return fmt.Errorf("error: Token not found.\n" +
				"Please run the command 'src auth login'.\n" +
				"Or set the environment variable SOURCECRAFT_TOKEN (for CI/CD).")
		}

		apiClient = api.NewClient("https://api.sourcecraft.tech", token)

		return nil
	},
}

func Execute() {
	cobra.OnInitialize(initConfig)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")

		home, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(home, ".config", "src")
			viper.AddConfigPath(configPath)
		}

		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, " ", viper.ConfigFileUsed())
	} else {

		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}

	}

}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "$HOME/.config/src"
	}
	return filepath.Join(home, ".config", "src")
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (example: C:\\Users\\User\\.config\\src\\config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Detailed log output")
}
