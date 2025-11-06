package cmd

import (
	"encoding/json"
	"fmt"
	"queueCtl/internal/config"
	"strconv"

	"github.com/spf13/cobra"
)

func ConfigCmd(cfg *config.Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}


	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value (max-retries, backoff-base)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			switch key {
			case "max-retries":
				i, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid value for max-retries: %s", value)
				}
				cfg.MaxRetries = i
			case "backoff-base":
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("invalid value for backoff-base: %s", value)
				}
				cfg.BackoffBase = f
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if err := config.SaveConfig(cfg); err != nil {
				return err
			}

			fmt.Printf("%s = %s\n", key, value)
			return nil
		},
	}

	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(setCmd)
	return configCmd
}