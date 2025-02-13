#!/usr/bin/env bash

# cd into the root just in case someone forgot to do it
PATH_RE="$(pwd)"
cd "$(git rev-parse --show-toplevel)"

# Define the file to update
TARGET_FILE="README.md"

# Check if the README.md file exists
if [[ ! -f "$TARGET_FILE" ]]; then
    echo "Error: $TARGET_FILE not found. Please ensure the file exists in the current directory."
    exit 1
fi

# Run `go run . -help` and capture both stdout and stderr
HELP_OUTPUT=$(go run . -help 2>&1)

# Use awk to replace the content between the markdown comments
awk -v help_output="$HELP_OUTPUT" '
BEGIN { inside_block = 0 }
/\[\/\/\]: # \(BEGIN HELPINFO\)/ {
    print; 
    print "```bash"; 
    print help_output; 
    print "```"; 
    inside_block = 1; 
    next
}
/\[\/\/\]: # \(END HELPINFO\)/ { 
    inside_block = 0 
}
!inside_block' "$TARGET_FILE" > "${TARGET_FILE}.tmp" && mv "${TARGET_FILE}.tmp" "$TARGET_FILE"

# Notify the user
echo "Updated $TARGET_FILE with the latest help information."

# Go back to where pwd told the user was
cd $PATH_RE