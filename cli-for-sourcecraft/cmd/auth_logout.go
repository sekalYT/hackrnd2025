// cmd/auth_logout.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zalando/go-keyring"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logging Out (Deleting a Stored Token)",
	Long:  `Removes the stored personal access token from the configuration file.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := keyring.Delete(keyringServiceName, keyringTokenUser)
		if err != nil {
			if err != keyring.ErrNotFound {
				fmt.Printf("Warning: Failed to remove token from keyring: %v\n", err)
			}
		}

		viper.Set("token", nil)

		configFileUsed := viper.ConfigFileUsed()
		if configFileUsed == "" {
			err := viper.ReadInConfig()
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println("Configuration file not found. The local token (if any) has been deleted.")
				return nil
			}
			configFileUsed = viper.ConfigFileUsed()
		}

		if configFileUsed != "" {
			err := viper.WriteConfig()
			if err != nil {
				return fmt.Errorf("configuration update failed '%s': %w", configFileUsed, err)
			}
			fmt.Println("The token has been successfully removed from the configuration file and/or OS keyring.")
			fmt.Println("Path to the configuration file:", configFileUsed)
		} else {
			fmt.Println("The token has been removed from OS keyring.")
		}

		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}
