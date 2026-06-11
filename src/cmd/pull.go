package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the latest world diff from the backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pull")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
