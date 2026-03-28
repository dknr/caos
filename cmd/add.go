package cmd

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/dknr/caos/client"
	"github.com/spf13/cobra"
)

// AddCmd is the `caos add` command.
var AddCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add data to the store from stdin or a file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// For simplicity, we'll connect to a local server.
		// In a real application, we might want to make the server address configurable.
		c := client.NewClient("http://localhost:31923")

		var reader io.Reader
		if len(args) == 0 || args[0] == "-" {
			reader = os.Stdin
		} else {
			file, err := os.Open(args[0])
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			reader = file
		}

		// Store the data
		addr, err := c.Add(context.Background(), reader, "application/octet-stream")
		if err != nil {
			log.Fatal(err)
		}

		var displayName string
		if len(args) == 0 || args[0] == "-" {
			displayName = "stdin"
		} else {
			displayName = args[0]
		}
		log.Printf("Added %s with address %s", displayName, addr)
	},
}