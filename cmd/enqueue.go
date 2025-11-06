package cmd

import (
	"encoding/json"
	"fmt"
	"queueCtl/internal/config"
	"queueCtl/internal/model"
	"queueCtl/internal/storage"
	"time"
	"github.com/spf13/cobra"
)


func EnqueueCmd(store *storage.Store, cfg *config.Config) *cobra.Command {
	var EnqueueCmd = &cobra.Command{
		Use: "enqueue <job(json)>",
		Short: "adds the job to the queue",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error{
			var job model.Job
			if err := json.Unmarshal([]byte(args[0]), &job); err != nil {
				return fmt.Errorf("invalid job JSON: %w", err)
			}
			
			if job.ID == "" || job.Command == "" {
				return fmt.Errorf("job 'id' or 'command' is empty")
			}

			now := time.Now()
			job.State = "pending"
			job.CreatedAt = now
			job.UpdatedAt = now
			job.NextRunAt = now

			if job.MaxRetries == 0{
				job.MaxRetries = cfg.MaxRetries
			}

			if err:=store.CreateJob(&job); err!=nil{
				return fmt.Errorf("failed to enqueue job: %v", err)
			}
			fmt.Println("Job enqueued.")
			return  nil
		},
	}
	return EnqueueCmd
}