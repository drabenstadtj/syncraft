package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a world diff to the backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("push")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
