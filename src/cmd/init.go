package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [world]",
	Short: "Initialize syncraft in a Minecraft world directory",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := bufio.NewScanner(os.Stdin)

		var worldDir string
		if len(args) == 1 {
			dir, err := findWorldDir(args[0])
			if err != nil {
				return err
			}
			worldDir = dir
		} else {
			savesDir, err := config.DefaultSavesDir()
			if err != nil {
				return fmt.Errorf("find saves dir: %w", err)
			}
			worlds, err := listWorlds(savesDir)
			if err != nil {
				return fmt.Errorf("list worlds: %w", err)
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
			scanner.Scan() // consume newline
			worldDir = filepath.Join(savesDir, worlds[choice-1])
		}

		if config.IsInitialized(worldDir) {
			fmt.Print(".syncraft already exists here. Reinitialize? [y/N]: ")
			scanner.Scan()
			if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
				return nil
			}
		}

		fmt.Print("Server URL (leave blank to set later): ")
		scanner.Scan()
		server := strings.TrimSpace(scanner.Text())

		snapshotDir := config.SnapshotDir(worldDir)
		fmt.Println("Snapshotting...")
		copied, err := snapshotWorld(worldDir, snapshotDir)
		if err != nil {
			return fmt.Errorf("snapshot: %w", err)
		}

		wc := &config.WorldConfig{Server: server}
		if err := wc.Save(worldDir); err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		fmt.Printf("initialized %s — snapshotted %d file(s)\n", filepath.Base(worldDir), copied)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

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
		if _, err := os.Stat(filepath.Join(savesDir, e.Name(), "level.dat")); err == nil {
			worlds = append(worlds, e.Name())
		}
	}
	return worlds, nil
}
