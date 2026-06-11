package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

// trackedPatterns lists the file patterns to snapshot, relative to the world root.
var trackedPatterns = []string{
	"level.dat",
	"playerdata/*.dat",
	"region/r.*.*.mca",
	"entities/r.*.*.mca",
	"poi/r.*.*.mca",
}

var initCmd = &cobra.Command{
	Use:   "init <name> <world-dir>",
	Short: "Register a world and take an initial snapshot",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, worldDir := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if _, exists := cfg.Worlds[name]; exists {
			return fmt.Errorf("world %q already registered — remove it first", name)
		}

		snapshotDir, err := config.SnapshotDir(name)
		if err != nil {
			return err
		}

		copied, err := snapshotWorld(worldDir, snapshotDir)
		if err != nil {
			return fmt.Errorf("snapshot: %w", err)
		}

		cfg.Worlds[name] = config.World{Path: worldDir, Snapshot: snapshotDir}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		fmt.Printf("registered %q — snapshotted %d file(s)\n", name, copied)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// snapshotWorld copies all tracked files from worldDir into snapshotDir,
// preserving subdirectory structure.
func snapshotWorld(worldDir, snapshotDir string) (int, error) {
	var count int
	for _, pattern := range trackedPatterns {
		matches, err := filepath.Glob(filepath.Join(worldDir, pattern))
		if err != nil {
			return 0, err
		}
		for _, src := range matches {
			rel, err := filepath.Rel(worldDir, src)
			if err != nil {
				return 0, err
			}
			dst := filepath.Join(snapshotDir, rel)
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				return 0, err
			}
			if err := copyFile(src, dst); err != nil {
				return 0, fmt.Errorf("copy %s: %w", rel, err)
			}
			count++
		}
	}
	return count, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
