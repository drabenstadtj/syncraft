package cmd

import (
	"fmt"
	"os"

	"github.com/drabenstadtj/syncraft/src/internal/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <world-before> <world-after> <output.syncdiff>",
	Short: "Generate a diff between two world region directories",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirA, dirB, outPath := args[0], args[1], args[2]

		diffs, err := diff.DiffWorld(dirA, dirB)
		if err != nil {
			return fmt.Errorf("diff: %w", err)
		}

		if len(diffs) == 0 {
			fmt.Println("no changes")
			return nil
		}

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		defer f.Close()

		if err := diff.Encode(f, diffs); err != nil {
			return fmt.Errorf("encode: %w", err)
		}

		total := 0
		for _, d := range diffs {
			total += len(d.Chunks)
		}
		fmt.Printf("wrote %d changed chunk(s) across %d region(s) to %s\n", total, len(diffs), outPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
