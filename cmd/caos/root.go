package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "caos",
	Short: "Caos — Content-Addressed Object Store",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
