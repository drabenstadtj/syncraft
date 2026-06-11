package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/drabenstadtj/syncraft/src/internal/diff"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <world>",
	Short: "Push world changes to the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		worldDir, err := findWorldDir(args[0])
		if err != nil {
			return err
		}

		wc, err := config.LoadWorldConfig(worldDir)
		if err != nil {
			return fmt.Errorf("world not initialized — run: syncraft init %s", args[0])
		}
		if wc.Server == "" {
			return fmt.Errorf("no server configured — run: syncraft remote %s <url>", args[0])
		}

		snapshotDir := config.SnapshotDir(worldDir)
		wd, err := diff.DiffWorld(snapshotDir, worldDir)
		if err != nil {
			return fmt.Errorf("diff: %w", err)
		}
		if len(wd.Regions) == 0 && len(wd.Files) == 0 {
			fmt.Println("nothing to push")
			return nil
		}

		var buf bytes.Buffer
		if err := diff.Encode(&buf, wd); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
		diffSize := buf.Len()

		resp, err := http.Post(wc.Server, "application/octet-stream", &buf)
		if err != nil {
			return fmt.Errorf("post to server: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
		}

		if _, err := snapshotWorld(worldDir, snapshotDir); err != nil {
			return fmt.Errorf("update snapshot: %w", err)
		}

		totalChunks := 0
		for _, d := range wd.Regions {
			totalChunks += len(d.Chunks)
		}
		fmt.Printf("pushed %d region(s), %d chunk(s), %d file(s) (%s)\n",
			len(wd.Regions), totalChunks, len(wd.Files), formatSize(diffSize))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
