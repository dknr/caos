package main

import (
	"fmt"
	"log/slog"
	"os"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	pushCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push [paths...]",
	Short: "Upload a directory tree as a path object",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		c := client.New(server)

		for _, arg := range args {
			pathAddr, count, err := c.PushDir(arg)
			if err != nil {
				slog.Error("Push failed", "path", arg, "error", err)
				os.Exit(1)
			}
			fmt.Printf("%s/path/%s (%d files)\n", server, pathAddr[:8], count)
		}
	},
}
