# queuectl: A Go-based CLI for a Persistent Job Queue System

`queuectl` is a command-line interface for managing a background job queue system. It is built in Go and designed as a minimal, production-grade service.

This system supports persistent job storage, a concurrent worker pool, automatic job retries with exponential backoff, and a Dead Letter Queue (DLQ).

## Features

* **Persistent Job Storage:** Job state is persisted in an embedded SQLite database, ensuring data integrity across application restarts.
* **Concurrent Worker Pool:** The system utilizes a goroutine-based worker pool to process multiple jobs in parallel.
* **Atomic Job Locking:** Jobs are leased to workers using database transactions, preventing race conditions and ensuring that a job is only processed by one worker at a time.
* **Retry Mechanism:** Failed jobs are automatically retried with a configurable exponential backoff delay.
* **Dead Letter Queue (DLQ):** Jobs that exhaust their retry attempts are moved to a DLQ for manual inspection and intervention.
* **Graceful Shutdown:** The worker process listens for `SIGINT` and `SIGTERM` signals, allowing in-progress jobs to finish before the process exits.
* **CLI Interface:** All system interactions are handled through a clean, `cobra`-based CLI.

---

## Setup Instructions

### 1. Prerequisites

* **Go:** Version 1.18 or higher.
* **C Compiler:** The `mattn/go-sqlite3` driver requires CGo.
    * **macOS:** `xcode-select --install`
    * **Linux (Debian/Ubuntu):** `sudo apt install build-essential`
    * **Windows:** A valid `gcc` installation, such as one provided by [TDM-GCC](http://tdm-gcc.tdragon.net/) or `mingw-w64`.

### 2. Build

Clone the repository and compile the binary:

```bash
# 1. Clone the repository
git clone [https://github.com/](https://github.com/)[your-username]/[your-repo-name].git
cd [your-repo-name]

# 2. Build the executable
go build -o queuectl.exe
```
## 3. Usage

### Config Commands
```bash
# Values that can be updated: data-dir, backoff-base, max-retries
./queuectl config set backoff-base 3

# shows current config values
./queuectl config show
```
- output
```bash
Updated. backoff-base = 3

{
  "data_dir": "./db",
  "max_retries": 4,
  "backoff_base": 3
}
```

### Enqueue a New Job
```bash
# Enqueue a simple job
./queuectl enqueue '{"id":"job-1", "command":"echo Hello World"}'

# Enqueue a job that will fail, triggering retries
./queuectl enqueue '{"id":"job-2", "command":"exit 1"}'
```
 
### Start the Worker Pool
```bash
# Start a pool of 3 workers
./queuectl worker start --count 3
```
- output:
```bash
$ ./queuectl worker start --count 3
2025/11/07 16:09:56 Starting 3 worker(s)...
2025/11/07 16:09:56 Use 'worker stop' command in different terminal to shutdown the workers.
2025/11/07 16:09:56 Worker 1: Starting
2025/11/07 16:09:56 Worker 2: Starting
2025/11/07 16:09:56 Worker 3: Starting
2025/11/07 16:09:57 Worker 3: Processing job job-1 (command: echo Hello World)
2025/11/07 16:09:57 Worker 3: Job job-1 completed successfully
```

### Stop the Worker Pool
```bash
./queuectl worker stop
```
### Check System Status
```bash
./queuectl status
```
- output
```bash
./queuectl.exe status
--- Job Queue Status ---
completed:      2
failed:         1

--- Worker Status ---
Workers:        3 started at: 2025-11-07 17:41:20.2522872 +0530 IST 
PID of worker pool: 27960
```
### List Jobs by State
job states: pending, processing, completed, failed, dead
```bash
./queuectl list --state failed
```
-output
```bash
./queuectl.exe list --state pending
--- Jobs in 'pending' state ---
ID              Command         Attempts
job-2           echo Hello World2               0
job-fail                ech Hello World2                0
```

### Manage the Dead Letter Queue (DLQ)
```bash
# List all jobs in the DLQ
./queuectl dlq list

# Re-queue a specific job (moves it from 'dead' to 'pending')
./queuectl dlq retry job-fail
```
- output
```bash
$ ./queuectl dlq list
--- Jobs in DLQ ---

--- Job 1 ---
ID:             job-2
Command:        exit 1
Attempts:       4
Last Updated:   2025-11-07T19:20:27+05:30
Last Output:    (empty)               4
$ ./queuectl.exe dlq retry job-fail
2025/11/07 17:43:19 Job job-fail moved from DLQ to 'pending' state.
```
## Architecture Overview 
1. **CLI (Cobra)**: The queuectl binary, built with cobra, acts as the user-facing controller. It's a short-lived process that writes commands (like enqueue or dlq retry) to the database and then exits.

2. **Storage (SQLite)**: The single source of truth is a SQLite database (queue.db). This was chosen over a simple file to safely manage concurrent access from multiple workers, preventing race conditions.

    - WAL Mode: The database runs in Write-Ahead Log (WAL) mode to allow high-concurrency reads and writes, which is critical for a multi-worker setup.

3. **Job Lifecycle**: Jobs transition through a well-defined state machine:

    - pending: A job is enqueued and waiting.

    - processing: A worker has locked the job.

    - completed: The job's command exited with code 0.

     - failed: The command exited with a non-zero code. The job is scheduled for a retry using exponential backoff.

    - dead: The job exhausted its max_retries and is moved to the Dead Letter Queue.

4. **Worker Pool (Goroutines)**: The queuectl worker start --count N command starts one OS process, which in turn spawns N goroutines (a worker pool).

    - Polling: Each worker goroutine runs an independent loop, polling the database to find an available job.

    - Atomic Locking: To prevent two workers from grabbing the same job, the FindAndLockJob function executes a SELECT and UPDATE within a database transaction. This makes the "leasing" of a job an atomic operation.

    - Graceful Shutdown: The worker start process listens for SIGINT and SIGTERM signals. Upon receiving one, it uses a Go context to signal all workers to finish their current job and exit. A sync.WaitGroup ensures the main process doesn't exit until all workers are done.

    - Stale Job Recovery: The worker query is designed to recover "orphaned" jobs. If a job is in the processing state for too long (e.g., 5 minutes), it's considered stale (due to a worker crash), and another worker will pick it up.

 5. **Inter-Process Communication (IPC)**: A simple IPC mechanism is used for the status and stop commands.

    - worker start writes its Process ID (PID) to a worker.status file.

    - status reads this file to report the number of active workers.

    - stop reads the PID from this file and sends a SIGTERM signal (using taskkill or syscall.SIGINT) to the worker process, triggering the graceful shutdown.

## Assumptions & Trade-offs
- **Queue Mechanism**: Implemented a database polling model where workers run a SELECT loop. This is simple but less efficient at scale than a true pub/sub system (like RabbitMQ)
- **Inter-Process Communication (IPC)**: Used a PID/status file for the stop and status commands. This is a simple IPC method but is a brittle solution as it can become stale if the worker process crashes, requiring manual cleanup.
- **Storage Engine**: Using SQLite provides an embedded, zero-dependency store. This trades the high write-concurrency of a networked database (e.g., Postgres) for simplicity.

- **Job Execution**: Using os/exec with sh -c is flexible but assumes all job commands are trusted. It provides no sandboxing.

- **Observability**: Capturing stdout/stderr to a DB output column aids DLQ debugging but is less scalable than streaming logs to a dedicated service.
## Testing Instructions
A shell script, test.sh, is included to provide an end-to-end validation of the core application flow.
The script will:

1. Build the queuectl binary.

2. Clear any old database files.

3. Enqueue a successful job, a failing job, and a long-running job.

4. Start the worker pool in the background and log its output.

5. Wait 15 seconds to allow the retry/backoff mechanism to run.

6. Gracefully stop the worker pool.

7. Check the final database status (expecting 2 completed, 1 dead).

8. Verify the failed job is in the DLQ and then retry it.

9. Print the worker logs for review.

### Running the Test Script
```bash
# 1. Make the script executable (one-time command)
chmod +x test.sh

# 2. Run the script
./test.sh
```