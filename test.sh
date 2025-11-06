#!/bin/bash

# --- Test Script for queuectl ---
# This script validates the core flows of the application.

# Set a color for sections
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}--- Step 1: Building the 'queuectl' binary ---${NC}"
go build -o queuectl.exe .
if [ $? -ne 0 ]; then
    echo "Build failed. Exiting."
    exit 1
fi

echo "Build successful."

# Ensure the DB directory exists
mkdir -p ./db

echo -e "\n${GREEN}--- Step 2: Cleaning the queue (deleting old DB) ---${NC}"
rm -f ./db/queue.db
rm -f ./db/worker.status
echo "Old database and status files removed."

echo -e "\n${GREEN}--- Step 3: Enqueuing a successful job, a failing job, and a long job ---${NC}"
./queuectl.exe enqueue '{"id":"job-success", "command":"echo Hello from the successful job"}'
./queuectl.exe enqueue '{"id":"job-fail", "command":"exit 1"}'
./queuectl.exe enqueue '{"id":"job-long", "command":"sleep 4 && echo Long job done"}'
echo "Jobs enqueued."

echo -e "\n${GREEN}--- Step 4: Checking queue status (3 jobs pending) ---${NC}"
./queuectl.exe status

echo -e "\n${GREEN}--- Step 5: Starting 3 workers in the background ---${NC}"
# Start the workers and send their output to a log file
./queuectl.exe worker start --count 3 > workers.log 2>&1 &
WORKER_PID=$!
echo "Workers started in background with PID $WORKER_PID. Logging to 'workers.log'."

echo -e "\n${GREEN}--- Step 6: Waiting for jobs to process (approx 15 seconds) ---${NC}"
echo "This will test success, retries, backoff, and DLQ..."
sleep 15 # Give time for:
# job-success (1s)
# job-long (4s)
# job-fail (1s + 2s backoff + 1s + 4s backoff + 1s = 9s)

echo "Done waiting."

echo -e "\n${GREEN}--- Step 7: Stopping the workers gracefully ---${NC}"
./queuectl.exe worker stop
sleep 2 # Give workers time to fully stop

echo -e "\n${GREEN}--- Step 8: Final Queue Status (Expected: 2 completed, 1 dead) ---${NC}"
./queuectl.exe status

echo -e "\n${GREEN}--- Step 9: Verifying the Dead Letter Queue (Expected: job-fail) ---${NC}"
./queuectl.exe dlq list

echo -e "\n${GREEN}--- Step 10: Retrying the failed job ---${NC}"
./queuectl.exe dlq retry job-fail
./queuectl.exe list --state pending

echo -e "\n${GREEN}--- Step 11: Reviewing worker logs ---${NC}"
echo "Displaying the last 20 lines of 'workers.log':"
tail -n 20 workers.log

echo -e "\n${GREEN}--- Test Complete ---${NC}"