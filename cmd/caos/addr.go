package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addrCmd)
}

var addrCmd = &cobra.Command{
	Use:   "addr [partial]",
	Short: "Resolve a partial address to full address(es)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := newClient()

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
