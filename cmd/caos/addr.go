package main

import (
	"fmt"
	"log/slog"
	"os"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addrCmd)
	addrCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
}

var addrCmd = &cobra.Command{
	Use:   "addr [partial]",
	Short: "Resolve a partial address to full address(es)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		c := client.New(server)

		addrs, err := c.ResolveAddr(args[0])
		if err != nil {
			slog.Error("Resolve failed", "error", err)
			os.Exit(1)
		}
		for _, a := range addrs {
			fmt.Println(a)
		}
	},
}
