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

var pullCmd = &cobra.Command{
	Use:   "pull <world>",
	Short: "Pull and apply the latest world diff from the server",
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

		resp, err := http.Get(wc.Server)
		if err != nil {
			return fmt.Errorf("fetch from server: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Println("no diff available on server")
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		diffSize := len(body)

		wd, err := diff.Decode(bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("decode: %w", err)
		}

		if err := diff.Patch(worldDir, wd); err != nil {
			return fmt.Errorf("patch: %w", err)
		}

		snapshotDir := config.SnapshotDir(worldDir)
		if _, err := snapshotWorld(worldDir, snapshotDir); err != nil {
			return fmt.Errorf("update snapshot: %w", err)
		}

		totalChunks := 0
		for _, d := range wd.Regions {
			totalChunks += len(d.Chunks)
		}
		fmt.Printf("applied %d region(s), %d chunk(s), %d file(s) (%s)\n",
			len(wd.Regions), totalChunks, len(wd.Files), formatSize(diffSize))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
