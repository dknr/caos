package main

import (
	"log/slog"
	"net/http"
	"os"

	"caos.one/caos/server"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntP("port", "p", 31923, "Port to listen on")
	serveCmd.Flags().StringP("root", "r", "", "Root directory for meta and data stores (default: /tmp/caos-<pid>)")
	serveCmd.Flags().StringP("home", "H", "/data/d10b49b4", "Redirect target for GET /")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the caos server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

		root, _ := cmd.Flags().GetString("root")
		if root == "" {
			root = "/tmp/caos-" + os.Getenv("PID")
			if root == "/tmp/caos-" {
				root = "/tmp/caos-" + "default"
			}
		}

		port, _ := cmd.Flags().GetInt("port")
		homePath, _ := cmd.Flags().GetString("home")
		srv := server.NewWithPort(root, port, homePath)

		if err := srv.Serve(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}

		slog.Info("Server stopped.")
	},
}
