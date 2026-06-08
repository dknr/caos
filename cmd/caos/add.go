package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Upload a file to the caos server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := newClient()

		path := args[0]
		f, err := os.Open(path)
		if err != nil {
			slog.Error("Failed to open file", "error", err)
			os.Exit(1)
		}
		defer f.Close()

		addr, err := c.AddData(f)
		if err != nil {
			slog.Error("Upload failed", "error", err)
			os.Exit(1)
		}
		fmt.Println(addr)
	},
}
