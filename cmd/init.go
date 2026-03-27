package cmd

import (
	"log"
    	"os"
    	"github.com/spf13/cobra"
)

// InitCmd is the `caos init` command.
var InitCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a new CAOS store directory structure",
    Run: func(cmd *cobra.Command, args []string) {
        // Define the store structure.
        dataPath := "caos-store/caos-datastore"
        metaPath := "caos-store/caos-metastore"
        dbFile := metaPath + "/caos-objs.db"

        if err := os.MkdirAll(dataPath, 0o755); err != nil {
            log.Fatalf("failed to create data directory: %v", err)
        }
        if err := os.MkdirAll(metaPath, 0o755); err != nil {
            log.Fatalf("failed to create meta directory: %v", err)
        }
        // Touch the SQLite file to ensure it exists.
        f, err := os.Create(dbFile)
        if err != nil {
            log.Fatalf("failed to create db file: %v", err)
        }
        f.Close()
        log.Printf("initialized store at %s", dbFile)
    },
}

// Register the init command.
func init() {
    // Import cobra to access the Command type.
    _ = cobra.Command{}
}
