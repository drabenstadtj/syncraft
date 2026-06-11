package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Generate a diff of world changes since the last sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("diff")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
