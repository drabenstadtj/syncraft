package cmd

import (
	"fmt"
	"os"

	"github.com/drabenstadtj/syncraft/src/internal/diff"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <world-region-dir> <input.syncdiff>",
	Short: "Apply a world diff to a region directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		worldDir, diffPath := args[0], args[1]

		f, err := os.Open(diffPath)
		if err != nil {
			return fmt.Errorf("open diff: %w", err)
		}
		defer f.Close()

		diffs, err := diff.Decode(f)
		if err != nil {
			return fmt.Errorf("decode: %w", err)
		}

		for _, d := range diffs {
			target := worldDir + "/" + d.Filename
			if err := diff.Patch(target, d); err != nil {
				return fmt.Errorf("patch %s: %w", d.Filename, err)
			}
			fmt.Printf("patched %s (%d chunk(s))\n", d.Filename, len(d.Chunks))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
