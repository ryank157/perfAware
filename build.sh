#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

echo "Building Go commands..."

# Get all subdirectories in the cmd directory
cmd_dirs=$(find cmd -maxdepth 1 -type d -not -path "cmd")

# Build each command in its respective directory
for dir in $cmd_dirs; do
    command_name=$(basename "$dir")  # Extract command name

    echo "Building command: $command_name"

    if [[ "$*" == *"--debug"* ]]; then
      go build -gcflags="all=-N -l" -o "./$command_name" "./$dir"
      echo "Debug build of command built successfully: $command_name"
    else
      go build -o "./$command_name" "./$dir"
      echo "Command built successfully: $command_name"
    fi

done

echo "All commands built successfully."
