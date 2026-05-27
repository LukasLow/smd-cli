# Agent Instructions for this project

ALWAYS use `smd` for EVERY command. Never run commands directly.

## Why?
Every command runs in an isolated container with:
- **Persistent session** — apt, pip, npm, files all survive across commands
- **Automatic .env protection** — sensitive files are blocked in temp mode
- **Pre-configured runtimes** — no need to install Node, Python, Go, etc.

## All commands go through smd
```
smd ls                 # List files
smd cat file.txt       # Read files
smd head file.txt      # View file start
smd python script.py   # Run Python
smd npm install        # Install npm packages
smd go build           # Build Go project
smd apt install curl   # Install system packages
smd git status         # Git operations
smd pip install flask  # Install Python packages
smd make               # Run Make
smd cargo build        # Build Rust project
```

## Tips
- **State persists** — first `smd <cmd>` creates a session container, subsequent commands reuse it
- **Temp mode** — use `smd -t <cmd>` for a truly ephemeral one-shot command
- **Cleanup** — `smd --close` destroys the session container
- **First run** — just run `smd <cmd>` and everything auto-detects

This ensures all commands execute inside the project's containerized environment.
