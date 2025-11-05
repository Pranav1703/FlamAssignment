package cmd

import (

	"log"

	"queueCtl/internal/config"
	"queueCtl/internal/storage"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "queueCtl",
	Short: "A cli-based job queue system",
}

func Execute(store *storage.Store, cfg *config.Config) {

	rootCmd.AddCommand(EnqueueCmd(store,cfg))
	rootCmd.AddCommand(ListCmd(store))
	rootCmd.AddCommand(StatusCmd(store))
	rootCmd.AddCommand(WorkerCmd(store, cfg))
	rootCmd.AddCommand(DlqCmd(store))
	rootCmd.AddCommand(ConfigCmd(cfg))

    if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
    }
}