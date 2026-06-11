package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var uninitCmd = &cobra.Command{
	Use:   "uninit <name>",
	Short: "Deregister a world and remove its snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		idx, err := config.LoadIndex()
		if err != nil {
			return fmt.Errorf("load index: %w", err)
		}
		worldDir, ok := idx.Worlds[name]
		if !ok {
			return fmt.Errorf("world %q not registered", name)
		}

		// remove .syncraft/ from the world folder
		dotDir := filepath.Join(worldDir, ".syncraft")
		if err := os.RemoveAll(dotDir); err != nil {
			return fmt.Errorf("remove world config: %w", err)
		}

		// remove snapshot
		snapshotDir, err := config.SnapshotDir(name)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(snapshotDir); err != nil {
			return fmt.Errorf("remove snapshot: %w", err)
		}

		// remove from index
		delete(idx.Worlds, name)
		if err := idx.Save(); err != nil {
			return fmt.Errorf("save index: %w", err)
		}

		fmt.Printf("removed %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninitCmd)
}
