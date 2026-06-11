package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/drabenstadtj/syncraft/src/internal/diff"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <name>",
	Short: "Push world changes to the server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		worldDir, snapshotDir, err := resolveWorld(name)
		if err != nil {
			return err
		}

		wc, err := config.LoadWorldConfig(worldDir)
		if err != nil {
			return fmt.Errorf("load world config: %w", err)
		}
		if wc.Server == "" {
			return fmt.Errorf("no server configured — run: syncraft remote %s <url>", name)
		}

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

		url := strings.TrimRight(wc.Server, "/") + "/worlds/" + name
		resp, err := http.Post(url, "application/octet-stream", &buf)
		if err != nil {
			return fmt.Errorf("post to server: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		if _, err := snapshotWorld(worldDir, snapshotDir); err != nil {
			return fmt.Errorf("update snapshot: %w", err)
		}

		totalChunks := 0
		for _, d := range wd.Regions {
			totalChunks += len(d.Chunks)
		}
		fmt.Printf("pushed %d region(s), %d chunk(s), %d file(s)\n",
			len(wd.Regions), totalChunks, len(wd.Files))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
