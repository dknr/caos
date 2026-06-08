package main

import (
	"caos.one/caos/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "caos",
	Short: "Caos — Content-Addressed Object Store",
}

func init() {
	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:31923", "Caos server URL")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "API key for write operations")

	viper.BindEnv("server", "CAOS_SERVER")
	viper.BindEnv("api-key", "CAOS_API_KEY")
}

func Execute() {
	cobra.OnInitialize(func() {
		viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
		viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	})
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

// newClient creates a client.Client configured from viper flags/env.
func newClient() *client.Client {
	c := client.New(viper.GetString("server"))
	if key := viper.GetString("api-key"); key != "" {
		c.SetAPIKey(key)
	}
	return c
}
