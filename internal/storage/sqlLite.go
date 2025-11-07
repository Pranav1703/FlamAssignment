package storage

import (
	"database/sql"
	"fmt"
	"log"
	"queueCtl/internal/model"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	Db *sql.DB
}

func (s *Store)Init() error{
	createJobTable :=`create table if not exists jobs(
		id text primary key,
		command text not null,
		state text not null default 'pending',
		attempts integer not null default 0,
		max_retries integer not null default 3,
		created_at DATETIME not null,
		updated_at DATETIME not null,
		next_run_at DATETIME,
		output text
	);`
	_, err := s.Db.Exec(createJobTable)
	return err
}


func NewStore(dbPath string)(*Store,error){
	db, err:= sql.Open("sqlite3",dbPath+"?_journal_mode=WAL")
	if err!= nil{
		return nil, err
	}

	if err:= db.Ping(); err!= nil{
		return nil,err
	}
	store := &Store{
		Db: db,
	}
	if err:= store.Init(); err!=nil{
		return nil,err
	}
	return store,nil
}

func (s *Store)CreateJob(job *model.Job) error{
	statement := `insert into jobs (
		id, command, state, attempts, max_retries, created_at, updated_at, next_run_at
		) Values (?,?,?,?,?,?,?,?);`
	_,err := s.Db.Exec(statement,job.ID,job.Command,job.State,job.Attempts,job.MaxRetries,job.CreatedAt,job.UpdatedAt,job.NextRunAt)
	if err!=nil{
		return err
	}
	return nil
}

func (s *Store)ListJobsByState(state string) ([]model.Job, error){
	statement := `select * from jobs where state=?`
	rows, err := s.Db.Query(statement,state)
	if err!=nil{
		return nil,err
	}
	defer rows.Close()

	var jobs []model.Job
	for rows.Next(){
		var job model.Job
		if err:= rows.Scan(
			&job.ID,
			&job.Command,
			&job.State,
			&job.Attempts,
			&job.MaxRetries,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.NextRunAt,
			&job.Output,
		); err!= nil{
			return nil,err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

//state -> count
func (s *Store)GetJobStats() (map[string]int, error){
	statement := `select state, count(*) from jobs group by state;`

	rows, err := s.Db.Query(statement)
	if err != nil{
		return nil, err
	}

	defer rows.Close()
	stateMap := make(map[string]int)

	for rows.Next(){
		var state string
		var count int

		if err :=rows.Scan(&state,&count);err!=nil{
			return nil,err
		}
		stateMap[state] = count
	}

	return stateMap,nil

}

func(s *Store)FindAndLock()(*model.Job,error){
	tx, err := s.Db.Begin()
	if err!=nil {
		return nil,err
	}
	defer tx.Rollback()
	statement := `select * from jobs where state =? or 
				(state = ? and next_run_at <= ?) or (state = ? and updated_at<= ?) order by created_at ASC LIMIT 1`

	row := tx.QueryRow(statement,model.StatePending, model.StateFailed,time.Now(),model.StateProcessing,time.Now())

	var job model.Job
	var nextRunAtStr sql.NullString
	var outputStr sql.NullString
	err = row.Scan(
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

	if err!=nil{
		if err==sql.ErrNoRows{
			return nil,nil
		}else{
			return nil,err
		}
	}

	if nextRunAtStr.Valid {
  		// This layout matches the format SQLite is storing: "YYYY-MM-DD HH:MM:SS.NNNNNNNNN+ZZ:ZZ"
		const sqliteTimeLayout = time.RFC3339Nano

		t, err := time.Parse(sqliteTimeLayout, nextRunAtStr.String)
        if err == nil {
			job.NextRunAt = t
        }
		if err!=nil{
			log.Println(err)
		}
    }
	if outputStr.Valid {
			job.Output = outputStr.String
	}

	updateStatement := `update jobs set 
					state = ?, 
					updated_at=?,
					attempts = ?
					where id = ?
					`
	now := time.Now()
	_, err = tx.Exec(updateStatement, model.StateProcessing, now, job.Attempts+1, job.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	job.State = model.StateProcessing
	job.UpdatedAt = now
	job.Attempts += 1
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