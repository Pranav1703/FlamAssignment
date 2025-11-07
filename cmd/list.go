package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"queueCtl/internal/config"
	"queueCtl/internal/database"

	"github.com/spf13/cobra"
)

func ListCmd(store *storage.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs by state",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get the --state flag
			state, _ := cmd.Flags().GetString("state")
			if state == "" {
				return fmt.Errorf("the --state flag is required")
			}

			jobs, err := store.ListJobsByState(state)
			if err != nil {
				return fmt.Errorf("failed to list jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Printf("No jobs found in state: %s\n", state)
				return nil
			}

			fmt.Printf("--- Jobs in '%s' state ---\n", state)
			fmt.Println("ID\t\tCommand\t\tAttempts")
			for _, job := range jobs {
				fmt.Printf("%s\t\t%s\t\t%d\n", job.ID, job.Command, job.Attempts)
			}
			return nil
		},
	}
	cmd.Flags().String("state", "", "Filter jobs by state (pending, processing, failed, dead, completed)")
	cmd.MarkFlagRequired("state")
	return cmd
}

func StatusCmd(store *storage.Store, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show a summary of job states",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := store.GetJobStats()
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			fmt.Println("--- Job Queue Status ---")
			if len(stats) == 0 {
				fmt.Println("No jobs in the queue.")
			}

			for state, count := range stats {
				fmt.Printf("%s: \t%d\n", state, count)
			}

			fmt.Println("\n--- Worker Status ---")
			statusPath := filepath.Join(cfg.DataDir, "worker.status")
			data, err := os.ReadFile(statusPath)

			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("Workers: \t0 (stopped)")
					return nil
				}
				return fmt.Errorf("could not read worker status: %w", err)
			}

			var status WorkerStatus
			if err := json.Unmarshal(data, &status); err != nil {
				return fmt.Errorf("could not parse worker status: %w", err)
			}

			fmt.Printf("Workers: \t%d started at: %v \nPID of worker pool: %d", status.Count, status.StartedAt, status.WorkerPoolPid)

			return nil
		},
	}
	return cmd
}
