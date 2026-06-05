package main

import (
	"fmt"
	"log/slog"
	"os"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	pullCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
	pullCmd.Flags().StringP("output", "o", ".", "Output directory for pulled files")
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull [addr...]",
	Short: "Pull a path object's files to disk",
	Long: `Download all files from one or more path objects to disk.

Each path object is resolved, its file listing is read, and each file is
written to the output directory with its relative path from the path index.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		outDir, _ := cmd.Flags().GetString("output")
		c := client.New(server)

		for _, addr := range args {
			if err := c.PullAddr(addr, outDir); err != nil {
				slog.Error("Pull failed", "addr", addr, "error", err)
				os.Exit(1)
			}
			fmt.Printf("%s -> %s\n", addr, outDir)
		}
	},
}
