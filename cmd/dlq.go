package cmd

import (
	"fmt"
	"log"
	"queueCtl/internal/model"
	"queueCtl/internal/storage"
	"time"

	"github.com/spf13/cobra"
)

func DlqCmd(store *storage.Store) *cobra.Command {
	dlqCmd := &cobra.Command{
		Use:   "dlq",
		Short: "Manage the Dead Letter Queue (DLQ)",
	}

	// --- 'dlq list' Subcommand ---
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all jobs in the DLQ",
		RunE: func(cmd *cobra.Command, args []string) error {
			jobs, err := store.ListJobsByState(model.StateDead)
			if err != nil {
				return fmt.Errorf("failed to list DLQ jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Println("Dead Letter Queue is empty.")
				return nil
			}

			fmt.Println("--- Jobs in DLQ ---")
			for i, job := range jobs {
				fmt.Printf("\n--- Job %d ---\n", i+1)
				fmt.Printf("ID: \t\t%s\n", job.ID)
				fmt.Printf("Command: \t%s\n", job.Command)
				fmt.Printf("Attempts: \t%d\n", job.Attempts)
				fmt.Printf("Last Updated: \t%s\n", job.UpdatedAt.Format(time.RFC3339))
				if job.Output != "" {
					fmt.Printf("Last Output: \n%s\n", job.Output)
				} else {
					fmt.Println("Last Output: \t(empty)")
				}
			}
			return nil
		},
	}

	// --- 'dlq retry' Subcommand ---
	retryCmd := &cobra.Command{
		Use:   "retry <job-id>",
		Short: "Retry a specific job from the DLQ",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := args[0]
			if err := store.RetryDeadJob(jobID); err != nil {
				return err
			}
			log.Printf("Job %s moved from DLQ to 'pending' state.", jobID)
			return nil
		},
	}

	dlqCmd.AddCommand(listCmd)
	dlqCmd.AddCommand(retryCmd)
	return dlqCmd
}