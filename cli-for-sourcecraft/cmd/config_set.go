// cmd/config_set.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set the value of the configuration parameter",
	Long:  `Sets or updates the value for the specified key in the configuration file.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if key == "token" {
			return fmt.Errorf("changing the token via 'config set' is not allowed. Use 'auth login'")
		}

		viper.Set(key, value)

		err := viper.WriteConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				configDir := getConfigDir()
				configPath := filepath.Join(configDir, "config.yaml")
				fmt.Printf("Configuration file not found, create a new one: %s\n", configPath)
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
				return fmt.Errorf("could not save configuration: %w", err)
			}
		}

		fmt.Printf("The '%s' parameter is set to '%s'\n", key, value)
		if viper.ConfigFileUsed() != "" {
			fmt.Println("Path to the configuration file:", viper.ConfigFileUsed())
		} else {
			fmt.Println("Path to the configuration file:", filepath.Join(getConfigDir(), "config.yaml"))
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
