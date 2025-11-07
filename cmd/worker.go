package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"queueCtl/internal/config"
	"queueCtl/internal/database"
	"queueCtl/internal/worker"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type WorkerStatus struct {
	WorkerPoolPid int       `json:"pid"`
	Count         int       `json:"count"`
	StartedAt     time.Time `json:"started_at"`
}

func WorkerCmd(store *storage.Store, cfg *config.Config) *cobra.Command {
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "Manage worker processes",
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start one or more worker processes",
		RunE: func(cmd *cobra.Command, args []string) error {
			count, _ := cmd.Flags().GetInt("count")

			log.Printf("Starting %d worker(s)...", count)
			log.Println("Use 'worker stop' command in different terminal to shutdown the workers.")
			status := WorkerStatus{
				WorkerPoolPid: os.Getpid(), // Get our own Process ID
				Count:         count,
				StartedAt:     time.Now(),
			}
			statusPath := filepath.Join(cfg.DataDir, "worker.status")

			data, err := json.Marshal(status)
			if err != nil {
				log.Printf("Warning: could not create status file: %v", err)
			} else {

				os.WriteFile(statusPath, data, 0644)
			}

			defer os.Remove(statusPath)

			// Set up graceful shutdown
			// This context will be canceled when an OS signal is received.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// A WaitGroup blocks until all workers have finished.
			var wg sync.WaitGroup

			// Start the workers
			for i := 1; i <= count; i++ {
				wg.Add(1) // Increment the WaitGroup counter
				w := worker.New(i, store, cfg)

				// Run the worker in a new goroutine
				// Pass 'ctx' so the worker knows when to shut down.
				go w.Run(ctx, &wg)
			}

			// Listen for shutdown signals (Ctrl+C)
			// This goroutine waits for a signal and calls 'cancel()'.
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				sig := <-sigCh
				log.Printf("Received signal: %v. Shutting down...", sig)
				cancel() // Cancel the context
			}()

			// Wait for all workers to exit
			wg.Wait()

			log.Println("All workers have shut down. Exiting.")
			return nil
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop running worker processes gracefully",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusPath := filepath.Join(cfg.DataDir, "worker.status")

			data, err := os.ReadFile(statusPath)
			if err != nil {
				if os.IsNotExist(err) {
					log.Println("Workers are not running (no status file found).")
					return nil
				}
				return fmt.Errorf("could not read worker status: %w", err)
			}
			var status WorkerStatus
			if err := json.Unmarshal(data, &status); err != nil {
				return fmt.Errorf("could not parse worker status: %w", err)
			}
			fmt.Println("Worker pool PID: ",status.WorkerPoolPid)
			if runtime.GOOS == "windows" {
				// This is an alternative to taskkill
				cmd := exec.Command("powershell", "-Command", "Stop-Process", "-Id", strconv.Itoa(status.WorkerPoolPid))
				if err := cmd.Run(); err != nil {
					log.Printf("Failed to stop process with taskkill: %v. Cleaning up status file.", err)
					os.Remove(statusPath)
					return err
				}
			} else {
				process, err := os.FindProcess(status.WorkerPoolPid)
				if err != nil {
					log.Printf("Could not find process with PID %d: %v", status.WorkerPoolPid, err)
					os.Remove(statusPath)
					return nil
				}
				
				if err := process.Signal(syscall.SIGINT); err != nil {
					log.Printf("Failed to send signal: %v. Cleaning up status file.", err)
					os.Remove(statusPath)
					return err
				}
			}

			log.Println("Signal sent. Workers should shut down gracefully.")
			return nil
		},
	}
	workerCmd.AddCommand(stopCmd)

	startCmd.Flags().Int("count", 1, "Number of workers to start")
	workerCmd.AddCommand(startCmd)

	return workerCmd
}

