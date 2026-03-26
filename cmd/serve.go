package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dknr/caos/internal/server"
	"github.com/dknr/caos/internal/store/datastore"
	"github.com/dknr/caos/internal/store/metastore"
	"github.com/spf13/cobra"
)

// ServeCmd is the `caos serve` command.
var ServeCmd = &cobra.Command{
	Use:   "serve [addr]",
	Short: "Start the CAOS server",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addr := ":31923"
		if len(args) > 0 {
			addr = args[0]
		}

		// Create stores
		dataStore := datastore.NewFilesystemDatastore("./caos-store/caos-datastore")
		metaStore, err := metastore.NewSQLiteMetaStore("./caos-store/caos-metastore/caos-objs.db")
		if err != nil {
			log.Fatal(err)
		}
		defer metaStore.Close()

		// Create server
		srv := server.NewServer(dataStore, metaStore, addr)

		// Start server in a goroutine
		go func() {
			log.Printf("Starting CAOS server on %s", addr)
			if err := srv.Start(); err != nil {
				log.Fatalf("Server failed to start: %v", err)
			}
		}()

		// Wait for interrupt signal to gracefully shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down server...")
		// TODO: Implement graceful shutdown
	},
}