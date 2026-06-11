package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Register a Minecraft world with syncraft",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		savesDir, err := config.DefaultSavesDir()
		if err != nil {
			return fmt.Errorf("find saves dir: %w", err)
		}

		worlds, err := listWorlds(savesDir)
		if err != nil {
			return fmt.Errorf("list worlds in %s: %w", savesDir, err)
		}
		if len(worlds) == 0 {
			return fmt.Errorf("no worlds found in %s", savesDir)
		}

		fmt.Printf("Minecraft saves: %s\n\n", savesDir)
		for i, w := range worlds {
			fmt.Printf("  [%d] %s\n", i+1, w)
		}
		fmt.Print("\nPick a world: ")

		var choice int
		if _, err := fmt.Scan(&choice); err != nil || choice < 1 || choice > len(worlds) {
			return fmt.Errorf("invalid choice")
		}
		worldDir := filepath.Join(savesDir, worlds[choice-1])

		// default name to folder name, allow override
		defaultName := worlds[choice-1]
		fmt.Printf("World name [%s]: ", defaultName)
		var name string
		fmt.Scan(&name)
		name = strings.TrimSpace(name)
		if name == "" {
			name = defaultName
		}

		// check not already registered
		idx, err := config.LoadIndex()
		if err != nil {
			return fmt.Errorf("load index: %w", err)
		}
		if _, exists := idx.Worlds[name]; exists {
			return fmt.Errorf("world %q already registered", name)
		}

		// snapshot
		snapshotDir, err := config.SnapshotDir(name)
		if err != nil {
			return err
		}
		copied, err := snapshotWorld(worldDir, snapshotDir)
		if err != nil {
			return fmt.Errorf("snapshot: %w", err)
		}

		fmt.Print("Server URL (leave blank to set later): ")
		var server string
		fmt.Scan(&server)
		server = strings.TrimSpace(server)

		// write .syncraft/config.json inside the world folder
		wc := &config.WorldConfig{Name: name, Server: server}
		if err := wc.Save(worldDir); err != nil {
			return fmt.Errorf("write world config: %w", err)
		}

		// register in global index
		idx.Worlds[name] = worldDir
		if err := idx.Save(); err != nil {
			return fmt.Errorf("save index: %w", err)
		}

		fmt.Printf("\ninitialized %q — snapshotted %d file(s)\n", name, copied)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// listWorlds returns subdirectory names in savesDir that contain a level.dat.
func listWorlds(savesDir string) ([]string, error) {
	entries, err := os.ReadDir(savesDir)
	if err != nil {
		return nil, err
	}
	var worlds []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		levelDat := filepath.Join(savesDir, e.Name(), "level.dat")
		if _, err := os.Stat(levelDat); err == nil {
			worlds = append(worlds, e.Name())
		}
	}
	return worlds, nil
}
