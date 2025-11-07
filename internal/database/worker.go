package storage

import (
	"database/sql"
	"fmt"
	"log"
	"queueCtl/internal/model"
	"time"
)

func (s *Store) FindAndLock() (*model.Job, error) {
	findSQL := `
	UPDATE jobs SET
		state = ?,
		updated_at = ?,
		attempts = attempts + 1
	WHERE id = (
		SELECT id FROM jobs
		WHERE
			state = ?
			OR
			(state = ? AND next_run_at <= ?)
			OR
			(state = ? AND updated_at <= ?)
		ORDER BY created_at ASC
		LIMIT 1
	)
	RETURNING id, command, state, attempts, max_retries, created_at, updated_at, next_run_at, output
	`
	const JobTimeout = 5 * time.Minute

	var job model.Job
	var nextRunAtStr sql.NullString
	var outputStr sql.NullString
	now := time.Now()

	err := s.Db.QueryRow(findSQL,
		model.StateProcessing, // SET state
		now,                    // SET updated_at
		
		model.StatePending,    // WHERE state = 'pending'
		model.StateFailed,     // OR state = 'failed'
		now,
		model.StateProcessing, // OR state = 'processing'
		now.Add(-JobTimeout),
	).Scan(
		&job.ID,
		&job.Command,
		&job.State,
		&job.Attempts,
		&job.MaxRetries,
		&job.CreatedAt,
		&job.UpdatedAt,
		&nextRunAtStr,
		&outputStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			if err.Error() == "database is locked" {
				return nil, nil // Not an error, just try again later
			}
			return nil, err
		}
	}

	if nextRunAtStr.Valid {
		// This layout matches the format SQLite is storing: "YYYY-MM-DD HH:MM:SS.NNNNNNNNN+ZZ:ZZ"
		const sqliteTimeLayout = time.RFC3339Nano

		t, err := time.Parse(sqliteTimeLayout, nextRunAtStr.String)
		if err == nil {
			job.NextRunAt = t
		}
		if err != nil {
			log.Println(err)
		}
	}
	if outputStr.Valid {
		job.Output = outputStr.String
	}

	return &job, nil
}

// UpdateJob saves all fields of a job after execution.
func (s *Store) UpdateJob(job *model.Job) error {
	updateSQL := `UPDATE jobs SET 
	                  state = ?, 
	                  attempts = ?, 
	                  updated_at = ?, 
	                  next_run_at = ?,
					  output = ?
	              WHERE id = ?`
	_, err := s.Db.Exec(updateSQL,
		job.State,
		job.Attempts,
		job.UpdatedAt,
		job.NextRunAt,
		job.Output,
		job.ID,
	)
	return err
}

func (s *Store) RetryDeadJob(jobID string) error {
	// We reset the state, attempts, and next_run_at time
	sql := `UPDATE jobs SET state = ?, attempts = 0, next_run_at = ?
	        WHERE id = ? AND state = ?`
	
	res, err := s.Db.Exec(sql,
		model.StatePending,
		time.Now(),
		jobID,
		model.StateDead,
	)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no job found with ID '%s' in the dead state", jobID)
	}
	
	return nil
}
