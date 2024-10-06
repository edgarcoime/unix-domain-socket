#!/bin/bash

# -----------------------------------------------------------------------------
# Script Name: run_binary.sh
# Description: Runs a specified binary concurrently n times with optional flags
#              and logs output.
# Usage: ./run_binary.sh [-n number_of_instances] <binary_path> [binary_args...]
# Default: n=1
# -----------------------------------------------------------------------------

# Function to display usage information
usage() {
    echo "Usage: $0 [-n number_of_instances] <binary_path> [binary_args...]"
    echo ""
    echo "  -n    Number of instances to run (default: 1)"
    echo "  <binary_path>    Path to the executable binary to run"
    echo "  [binary_args...] Optional arguments and flags to pass to the binary"
    echo ""
    echo "Example:"
    echo "  $0 -n 5 /path/to/client --flag1 value1 --flag2 value2"
    exit 1
}

# Default number of instances
n=1

# Initialize variables
BINARY_PATH=""
BINARY_ARGS=()

# Parse command-line options
while getopts ":n:" opt; do
    case ${opt} in
        n )
            # Validate that the provided argument is a positive integer
            if ! [[ "$OPTARG" =~ ^[1-9][0-9]*$ ]]; then
                echo "Error: -n requires a positive integer."
                usage
            fi
            n=$OPTARG
            ;;
        \? )
            echo "Error: Invalid Option -$OPTARG" >&2
            usage
            ;;
        : )
            echo "Error: Option -$OPTARG requires an argument." >&2
            usage
            ;;
    esac
done

# Shift parsed options away
shift $((OPTIND -1))

# Check if the binary path is provided
if [ $# -lt 1 ]; then
    echo "Error: Missing binary path."
    usage
fi

BINARY_PATH="$1"
shift  # Shift the binary path, remaining are binary_args

# Assign remaining arguments to BINARY_ARGS
BINARY_ARGS=("$@")

# Check if the binary exists and is executable
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary '$BINARY_PATH' does not exist."
    exit 1
fi

if [ ! -x "$BINARY_PATH" ]; then
    echo "Error: Binary '$BINARY_PATH' is not executable."
    exit 1
fi

# Resolve absolute path of the binary
ABS_BINARY_PATH="$(cd "$(dirname "$BINARY_PATH")" && pwd)/$(basename "$BINARY_PATH")"

# Log file path (same directory as the script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="$SCRIPT_DIR/log.txt"

# Create or truncate the log file
> "$LOG_FILE"

echo "Running '$ABS_BINARY_PATH' $n time(s) with arguments: ${BINARY_ARGS[*]}"
echo "Logging to $LOG_FILE"
echo "----------------------------------------" >> "$LOG_FILE"
echo "Run started at $(date)" >> "$LOG_FILE"
echo "----------------------------------------" >> "$LOG_FILE"

# Function to run a single instance of the binary and append output to log.txt
run_instance() {
    local instance_num=$1
    {
        echo "----- Instance $instance_num Started at $(date) -----"
        "$ABS_BINARY_PATH" "${BINARY_ARGS[@]}"
        local exit_code=$?
        echo "----- Instance $instance_num Ended at $(date) with exit code $exit_code -----"
        echo ""
    } >> "$LOG_FILE" 2>&1
}

# Launch 'n' instances of the binary concurrently
for ((i=1; i<=n; i++)); do
    run_instance "$i" &
done

# Wait for all background processes to finish
wait

echo "All $n instance(s) of '$ABS_BINARY_PATH' have completed. Check $LOG_FILE for details."
echo "Run ended at $(date)" >> "$LOG_FILE"
echo "----------------------------------------" >> "$LOG_FILE"

