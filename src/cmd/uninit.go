package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var uninitCmd = &cobra.Command{
	Use:   "uninit <world>",
	Short: "Remove syncraft from a world directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		worldDir, err := findWorldDir(args[0])
		if err != nil {
			return err
		}

		if !config.IsInitialized(worldDir) {
			return fmt.Errorf("world %q is not initialized", args[0])
		}

		dotDir := filepath.Join(worldDir, config.DotDir)
		if err := os.RemoveAll(dotDir); err != nil {
			return fmt.Errorf("remove .syncraft: %w", err)
		}

		fmt.Printf("removed .syncraft from %s\n", filepath.Base(worldDir))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninitCmd)
}
