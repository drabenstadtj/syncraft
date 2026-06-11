package cmd

import (
	"fmt"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var remoteCmd = &cobra.Command{
	Use:   "remote <world> <url>",
	Short: "Set the server URL for a world",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		worldDir, err := findWorldDir(args[0])
		if err != nil {
			return err
		}

		wc, err := config.LoadWorldConfig(worldDir)
		if err != nil {
			return fmt.Errorf("world not initialized — run: syncraft init %s", args[0])
		}

		wc.Server = args[1]
		if err := wc.Save(worldDir); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		fmt.Printf("remote set to %s\n", args[1])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
}
