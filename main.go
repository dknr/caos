package main

import (
	"context"
	"log"

	"github.com/charmbracelet/fang"
	"github.com/dknr/caos/cmd"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "caos",
		Short: "Content-Addressed Object Store",
	}

	// Add commands
	root.AddCommand(cmd.ServeCmd)
	root.AddCommand(cmd.AddCmd)
	root.AddCommand(cmd.GetCmd)
	root.AddCommand(cmd.HelpCmd)
	root.AddCommand(cmd.InitCmd)

	if err := fang.Execute(context.Background(), root); err != nil {
		log.Fatal(err)
	}
}