#!/bin/bash

# Check if an argument is provided
if [ -z "$1" ]; then
    echo "Error: Please provide a container name or ID."
    echo "Usage: ./clean_container_log.sh <container_name_or_id>"
    exit 1
fi

CONTAINER_NAME=$1

# 1. Get the path to the log file
LOG_PATH=$(docker inspect --format='{{.LogPath}}' "$CONTAINER_NAME" 2>/dev/null)

# Check if the container exists
if [ $? -ne 0 ]; then
    echo "Error: Container '$CONTAINER_NAME' not found."
    exit 1
fi

# Check if the log file exists
if [ -z "$LOG_PATH" ] || [ ! -f "$LOG_PATH" ]; then
    echo "Notice: Container '$CONTAINER_NAME' has no log file or it has already been removed."
    exit 0
fi

# 2. Calculate size before cleaning
BEFORE_SIZE=$(du -h "$LOG_PATH" | cut -f1)

# 3. Perform the cleanup
# We use 'truncate' to empty the file content while keeping the file descriptor open.
sudo truncate -s 0 "$LOG_PATH"

# 4. Output results
echo "Successfully cleaned container: $CONTAINER_NAME"
echo "Log path: $LOG_PATH"
echo "Size before: $BEFORE_SIZE"
echo "Size after:  0"
