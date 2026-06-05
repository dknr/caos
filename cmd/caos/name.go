package main

import (
	"fmt"
	"log/slog"
	"os"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nameCmd)
	nameCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
}

var nameCmd = &cobra.Command{
	Use:   "name [name] [addr]",
	Short: "Get or set a name alias",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		c := client.New(server)

		if len(args) == 1 {
			// Get name
			addr, err := c.GetName(args[0])
			if err != nil {
				slog.Error("Get name failed", "error", err)
				os.Exit(1)
			}
			fmt.Println(addr)
		} else {
			// Set name
			if err := c.SetName(args[0], args[1]); err != nil {
				slog.Error("Set name failed", "error", err)
				os.Exit(1)
			}
		}
	},
}
