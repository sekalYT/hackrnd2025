// cmd/config_get.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get Configuration Parameter Value",
	Long:  `Displays the value for the specified key from the configuration file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		if key == "token" {
			fmt.Println("<скрыто>")
			return nil
		}

		value := viper.Get(key)

		if !viper.IsSet(key) {
			configFile := viper.ConfigFileUsed()
			if configFile == "" {
				err := viper.ReadInConfig()
				if _, ok := err.(viper.ConfigFileNotFoundError); ok || err != nil {
					fmt.Printf("Configuration file not found. The '%s' parameter is not set.\n", key)
					return nil
				}
				configFile = viper.ConfigFileUsed()
				if !viper.IsSet(key) {
					fmt.Printf("The '%s' parameter was not found in the configuration.\n", key)
					fmt.Println("Path to the configuration file:", configFile)
					return nil
				}
				value = viper.Get(key)
			} else {
				fmt.Printf("The '%s' parameter was not found in the configuration.\n", key)
				fmt.Println("Path to the configuration file:", configFile)
				return nil
			}
		}

		fmt.Printf("%v\n", value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
