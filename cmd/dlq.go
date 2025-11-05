package cmd

import (
	"fmt"
	"log"
	"queueCtl/internal/model"
	"queueCtl/internal/storage"

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
			fmt.Println("ID\t\tCommand\t\tAttempts")
			for _, job := range jobs {
				fmt.Printf("%s\t%s\t\t%d\n", job.ID, job.Command, job.Attempts)
			}
			return nil
		},
	}

	// --- 'dlq retry' Subcommand ---
	retryCmd := &cobra.Command{
		Use:   "retry [job-id]",
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