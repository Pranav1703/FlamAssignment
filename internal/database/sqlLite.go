package storage

import (
	"database/sql"
	"queueCtl/internal/model"

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
