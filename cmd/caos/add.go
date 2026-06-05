package main

import (
	"fmt"
	"log/slog"
	"os"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
}

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Upload a file to the caos server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		c := client.New(server)

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
