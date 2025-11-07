package storage

import (
	"database/sql"
	"queueCtl/internal/model"
)

func (s *Store) ListJobsByState(state string) ([]model.Job, error) {
	statement := `select * from jobs where state=?`
	rows, err := s.Db.Query(statement, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []model.Job
	var outputStr sql.NullString
	for rows.Next() {
		var job model.Job
		if err := rows.Scan(
			&job.ID,
			&job.Command,
			&job.State,
			&job.Attempts,
			&job.MaxRetries,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.NextRunAt,
			&outputStr,
		); err != nil {
			return nil, err
		}
		if outputStr.Valid {
			job.Output = outputStr.String
		}	
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// state -> count
func (s *Store) GetJobStats() (map[string]int, error) {
	statement := `select state, count(*) from jobs group by state;`

	rows, err := s.Db.Query(statement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	stateMap := make(map[string]int)

	for rows.Next() {
		var state string
		var count int

		if err := rows.Scan(&state, &count); err != nil {
			return nil, err
		}
		stateMap[state] = count
	}

	return stateMap, nil

}
