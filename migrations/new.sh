#!/bin/bash

# Check if a keyword is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <keyword>"
  exit 1
fi

# Get the current timestamp in nanoseconds
timestamp=$(date +%s)

# Get the keyword from the first argument, removing any spaces or special characters
keyword=$(echo "$1" | tr -cd '[:alnum:]_')

# Create the filename
upFilename="${timestamp}_${keyword}.up.sql"
downFilename="${timestamp}_${keyword}.down.sql"

# Create the .sql file
touch "migrations/$upFilename"
touch "migrations/$downFilename"

# Notify the user
echo "Created SQL file: migrations/$upFilename"
echo "Created SQL file: migrations/$downFilename"
