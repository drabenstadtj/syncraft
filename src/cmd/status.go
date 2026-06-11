package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [world]",
	Short: "Show initialized worlds and their configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return printWorldStatus(args[0])
		}

		savesDir, err := config.DefaultSavesDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(savesDir)
		if err != nil {
			return fmt.Errorf("read saves dir: %w", err)
		}

		found := false
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			worldDir := filepath.Join(savesDir, e.Name())
			if config.IsInitialized(worldDir) {
				printWorldStatus(e.Name())
				found = true
			}
		}

		if !found {
			fmt.Println("no initialized worlds — run: syncraft init")
		}
		return nil
	},
}

func printWorldStatus(name string) error {
	worldDir, err := findWorldDir(name)
	if err != nil {
		return err
	}

	wc, err := config.LoadWorldConfig(worldDir)
	server := "(none)"
	if err == nil && wc.Server != "" {
		server = wc.Server
	}

	fmt.Printf("world:   %s\n", name)
	fmt.Printf("path:    %s\n", worldDir)
	fmt.Printf("server:  %s\n\n", server)
	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
