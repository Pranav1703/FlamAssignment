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

### Stop the Worker Pool
```bash
./queuectl worker stop
```

### Check System Status
```bash
./queuectl status
```

### List Jobs by State
```bash
./queuectl list --state failed
```

### Manage the Dead Letter Queue (DLQ)
```bash
# List all jobs in the DLQ
./queuectl dlq list

# Re-queue a specific job (moves it from 'dead' to 'pending')
./queuectl dlq retry job-2
```

## Testing Instructions
A shell script, test.sh, is included to provide an end-to-end validation of the core application flow.
The script will:

Build the queuectl binary.

Clear any old database files.

Enqueue a successful job, a failing job, and a long-running job.

Start the worker pool in the background and log its output.

Wait 15 seconds to allow the retry/backoff mechanism to run.

Gracefully stop the worker pool.

Check the final database status (expecting 2 completed, 1 dead).

Verify the failed job is in the DLQ and then retry it.

Print the worker logs for review.

### Running the Test Script
```bash
# 1. Make the script executable (one-time command)
chmod +x test.sh

# 2. Run the script
./test.sh
```