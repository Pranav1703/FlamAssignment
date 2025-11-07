package cmd

import (

	"log"

	"queueCtl/internal/config"
	"queueCtl/internal/database"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "queueCtl",
	Short: "A cli-based job queue system",
}

func Execute(store *storage.Store, cfg *config.Config) {

	rootCmd.AddCommand(EnqueueCmd(store,cfg))
	rootCmd.AddCommand(ListCmd(store))
	rootCmd.AddCommand(StatusCmd(store,cfg))
	rootCmd.AddCommand(WorkerCmd(store, cfg))
	rootCmd.AddCommand(DlqCmd(store))
	rootCmd.AddCommand(ConfigCmd(cfg))

    if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
    }
}