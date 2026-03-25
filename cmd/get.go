package cmd

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/dknr/caos/client"
	"github.com/spf13/cobra"
)

// GetCmd is the `caos get` command.
var GetCmd = &cobra.Command{
	Use:   "get <addr> [output]",
	Short: "Get data from the store by address and write to stdout or a file",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		addr := args[0]
		output := "-"
		if len(args) > 1 {
			output = args[1]
		}

		// Connect to the local server
		c := client.NewClient("http://localhost:31923")

		// Get the data
		reader, err := c.Get(context.Background(), addr)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		// Write to output
		var writer io.WriteCloser
		if output == "-" {
			writer = os.Stdout
		} else {
			file, err := os.Create(output)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			writer = file
		}

		if _, err := io.Copy(writer, reader); err != nil {
			log.Fatal(err)
		}
	},
}