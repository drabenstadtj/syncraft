package cmd

import (
	"fmt"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote <name> <url>",
	Short: "Set the server URL for a registered world",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], args[1]

		worldDir, _, err := resolveWorld(name)
		if err != nil {
			return err
		}

		wc, err := config.LoadWorldConfig(worldDir)
		if err != nil {
			return fmt.Errorf("load world config: %w", err)
		}

		wc.Server = url
		if err := wc.Save(worldDir); err != nil {
			return fmt.Errorf("save world config: %w", err)
		}

		fmt.Printf("server for %q set to %s\n", name, url)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}
