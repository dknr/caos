package main

import (
	"os"
	"path/filepath"

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
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to config file (default: ~/.config/caos/caos.yaml)")

	viper.SetDefault("server", "http://localhost:31923")
	viper.SetDefault("api-key", "")
	viper.SetDefault("port", 31923)
	viper.SetDefault("root", "")
	viper.SetDefault("home", "/data/d10b49b4")
	viper.SetDefault("config", "")

	viper.BindEnv("server", "CAOS_SERVER")
	viper.BindEnv("api-key", "CAOS_API_KEY")
	viper.BindEnv("port", "CAOS_PORT")
	viper.BindEnv("root", "CAOS_ROOT")
	viper.BindEnv("home", "CAOS_HOME")
}

func Execute() {
	cobra.OnInitialize(func() {
		viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
		viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
		viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

		if cfg := viper.GetString("config"); cfg != "" {
			viper.SetConfigFile(cfg)
		} else {
			viper.SetConfigName("caos")
			viper.SetConfigType("yaml")
			viper.AddConfigPath(filepath.Join(os.ExpandEnv("${HOME}"), ".config", "caos"))
		}

		viper.ReadInConfig() // error is non-fatal
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
