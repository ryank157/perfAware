#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

echo "Building Go commands..."

DEBUG_BUILD=false
ENABLE_TIMING=false

# Check if --debug flag is present
for arg in "$@"; do
  if [ "$arg" == "--debug" ]; then
    DEBUG_BUILD=true
  elif [ "$arg" == "--timing" ]; then
    ENABLE_TIMING=true
  fi
done

# Get all subdirectories in the cmd directory
cmd_dirs=$(find cmd -maxdepth 1 -type d -not -path "cmd")

# Build each command in its respective directory
for dir in $cmd_dirs; do
  command_name=$(basename "$dir")  # Extract command name
  echo "Building command: $command_name"

  ldflags=""
  if $ENABLE_TIMING; then
      ldflags="-X github.com/ryank157/perfAware/internal/timing.enableTimingStr=true"  #Just the -X flag
  fi

  if $DEBUG_BUILD; then
      go build -gcflags="all=-N -l" -ldflags="$ldflags" -o "./$command_name" "./cmd/$command_name"
    echo "Debug build of command built successfully: $command_name"
  else
    go build -ldflags="$ldflags" -o "./$command_name" "./cmd/$command_name"
    echo "Command built successfully: $command_name"
  fi
done

echo "All commands built successfully."
