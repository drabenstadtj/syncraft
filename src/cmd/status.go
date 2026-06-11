package cmd

import (
	"fmt"

	"github.com/drabenstadtj/syncraft/src/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show registered worlds and their configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := config.LoadIndex()
		if err != nil {
			return fmt.Errorf("load index: %w", err)
		}

		if len(idx.Worlds) == 0 {
			fmt.Println("no worlds registered — run: syncraft init")
			return nil
		}

		// filter to a single world if name given
		worlds := idx.Worlds
		if len(args) == 1 {
			path, ok := idx.Worlds[args[0]]
			if !ok {
				return fmt.Errorf("world %q not registered", args[0])
			}
			worlds = map[string]string{args[0]: path}
		}

		for name, worldDir := range worlds {
			wc, err := config.LoadWorldConfig(worldDir)
			server := "(none)"
			if err == nil && wc.Server != "" {
				server = wc.Server
			}

			fmt.Printf("world:    %s\n", name)
			fmt.Printf("path:     %s\n", worldDir)
			fmt.Printf("server:   %s\n", server)
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
