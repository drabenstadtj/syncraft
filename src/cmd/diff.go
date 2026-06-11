package cmd

import (
	"fmt"
	"os"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/drabenstadtj/syncraft/src/internal/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <name> <output.syncdiff>",
	Short: "Generate a diff between the current world and the last snapshot",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, outPath := args[0], args[1]

		worldDir, err := findWorldDir(name)
		if err != nil {
			return err
		}
		snapshotDir := config.SnapshotDir(worldDir)
		if err != nil {
			return err
		}

		wd, err := diff.DiffWorld(snapshotDir, worldDir)
		if err != nil {
			return fmt.Errorf("diff: %w", err)
		}

		if len(wd.Regions) == 0 && len(wd.Files) == 0 {
			fmt.Println("no changes")
			return nil
		}

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		defer f.Close()

		if err := diff.Encode(f, wd); err != nil {
			return fmt.Errorf("encode: %w", err)
		}

		totalChunks := 0
		for _, d := range wd.Regions {
			totalChunks += len(d.Chunks)
		}
		fmt.Printf("%d region(s), %d chunk(s), %d file(s) changed — wrote %s\n",
			len(wd.Regions), totalChunks, len(wd.Files), outPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
