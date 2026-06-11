package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/drabenstadtj/syncraft/src/internal/diff"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <name>",
	Short: "Pull and apply the latest world diff from the server",
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

		url := strings.TrimRight(wc.Server, "/") + "/worlds/" + name
		resp, err := http.Get(url)
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

		wd, err := diff.Decode(resp.Body)
		if err != nil {
			return fmt.Errorf("decode: %w", err)
		}

		if err := diff.Patch(worldDir, wd); err != nil {
			return fmt.Errorf("patch: %w", err)
		}

		if _, err := snapshotWorld(worldDir, snapshotDir); err != nil {
			return fmt.Errorf("update snapshot: %w", err)
		}

		totalChunks := 0
		for _, d := range wd.Regions {
			totalChunks += len(d.Chunks)
		}
		fmt.Printf("applied %d region(s), %d chunk(s), %d file(s)\n",
			len(wd.Regions), totalChunks, len(wd.Files))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
