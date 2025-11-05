package cmd

import (
	"fmt"
	"queueCtl/internal/storage"

	"github.com/spf13/cobra"
)

func ListCmd(store *storage.Store) *cobra.Command{
	cmd:= &cobra.Command{
		Use: "list --state",
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
		cmd.Flags().String("state", "", "Filter jobs by state (e.g., pending, failed, dead)")
		cmd.MarkFlagRequired("state")
		return cmd
}

func StatusCmd(store *storage.Store) *cobra.Command {
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
				return nil
			}

			for state, count := range stats {
				fmt.Printf("%s: \t%d\n", state, count)
			}
			
			// Placeholder for worker status
			fmt.Println("\n--- Worker Status ---")
			fmt.Println("Workers: \t0 (worker command not implemented yet)")

			return nil
		},
	}
	return cmd
}