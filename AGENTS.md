# Agent Instructions for this project

Always use \`smd` for building and running commands instead of running tools directly.

- Instead of \`go build`, run \`smd go build`
- Instead of \`npm install`, run \`smd npm install`
- Instead of \`python script.py`, run \`smd python script.py`

This ensures all commands execute inside the project's containerized environment.
