package main

import (
	"io"
	"log/slog"
	"os"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
	getCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
}

var getCmd = &cobra.Command{
	Use:   "get [addr]",
	Short: "Download data by address",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		c := client.New(server)

		rc, _, err := c.GetData(args[0])
		if err != nil {
			slog.Error("Get data failed", "error", err)
			os.Exit(1)
		}
		defer rc.Close()

		outPath, _ := cmd.Flags().GetString("output")
		if outPath != "" {
			f, err := os.Create(outPath)
			if err != nil {
				slog.Error("Failed to create output file", "error", err)
				os.Exit(1)
			}
			defer f.Close()
			io.Copy(f, rc)
		} else {
			io.Copy(os.Stdout, rc)
		}
	},
}
