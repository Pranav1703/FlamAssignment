package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"queueCtl/internal/config"
	"queueCtl/internal/storage"
	"queueCtl/internal/worker"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
)

func WorkerCmd(store *storage.Store, cfg *config.Config) * cobra.Command{
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "Manage worker processes",
	}

	startCmd := &cobra.Command{
		Use: "start",
		Short: "Start one or more worker processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the --count flag
			count, _ := cmd.Flags().GetInt("count")

			log.Printf("Starting %d worker(s)...", count)
			log.Println("Press Ctrl+C to shut down gracefully.")

			// 1. Set up graceful shutdown
			// This context will be canceled when an OS signal is received.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// A WaitGroup blocks until all workers have finished.
			var wg sync.WaitGroup

			// 2. Start the workers
			for i := 1; i <= count; i++ {
				wg.Add(1) // Increment the WaitGroup counter
				w := worker.New(i, store, cfg)
				
				// Run the worker in a new goroutine
				// Pass 'ctx' so the worker knows when to shut down.
				go w.Run(ctx, &wg)
			}

			// 3. Listen for shutdown signals (Ctrl+C)
			// This goroutine waits for a signal and calls 'cancel()'.
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				sig := <-sigCh
				log.Printf("Received signal: %v. Shutting down...", sig)
				cancel() // Cancel the context
			}()

			// 4. Wait for all workers to exit
			wg.Wait()

			log.Println("All workers have shut down. Exiting.")
			return nil
		},
		
	}

	startCmd.Flags().Int("count", 1, "Number of workers to start")
	workerCmd.AddCommand(startCmd)

	return workerCmd
}