package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"caos.one/caos/client"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.Flags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
	tagCmd.Flags().BoolP("delete", "d", false, "Delete the tag")
}

var tagCmd = &cobra.Command{
	Use:   "tag [addr] [key] [value]",
	Short: "Get, set, or delete a tag on an address",
	Args:  cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		c := client.New(server)
		del, _ := cmd.Flags().GetBool("delete")

		addr := args[0]
		key := args[1]

		if del {
			if err := c.DelTag(addr, key); err != nil {
				slog.Error("Delete tag failed", "error", err)
				os.Exit(1)
			}
			return
		}

		if len(args) == 2 {
			// Get tag
			val, err := c.GetTag(addr, key)
			if err != nil {
				if strings.Contains(err.Error(), "404") {
					fmt.Println("(not found)")
					os.Exit(0)
				}
				slog.Error("Get tag failed", "error", err)
				os.Exit(1)
			}
			fmt.Println(val)
		} else {
			// Set tag
			if err := c.SetTag(addr, key, args[2]); err != nil {
				slog.Error("Set tag failed", "error", err)
				os.Exit(1)
			}
		}
	},
}
