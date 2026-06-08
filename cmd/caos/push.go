package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push [paths...]",
	Short: "Upload a directory tree as a path object",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := newClient()

		for _, arg := range args {
			pathAddr, count, err := c.PushDir(arg)
			if err != nil {
				slog.Error("Push failed", "path", arg, "error", err)
				os.Exit(1)
			}
			fmt.Printf("%s/path/%s (%d files)\n", c.Base(), pathAddr[:8], count)
		}
	},
}
