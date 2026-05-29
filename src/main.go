package main

import (
	"fmt"
	"os"
	"strings"
)

const version = "0.2.1"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printInfo()
		return
	}

	switch args[0] {
	case "-h", "--help":
		printHelp()
		return
	case "--version":
		fmt.Printf("smd version %s\n", version)
		return
	case "--agentmd":
		generateAgentMD()
		return
	case "-i", "--init":
		_, err := TemplateInitInteractive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nRun 'smd --open' for live mode or 'smd <command>' for session mode.\n")
		return
	case "-o", "--open":
		config, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		if config == nil {
			config, err = InteractiveInit()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}
		}
		err = RunLiveMode(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	case "--close":
		config, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		if config == nil {
			fmt.Println("No smd.toml found. Nothing to clean.")
			return
		}
		err = DestroySessionContainer(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for -t / --temp flag
	tempMode := false
	cmdStart := 0
	if args[0] == "-t" || args[0] == "--temp" {
		tempMode = true
		cmdStart = 1
		if len(args) < 2 {
			fmt.Println("Error: -t/--temp requires a command")
			os.Exit(1)
		}
	}

	command := args[cmdStart:]

	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if config == nil {
		cmd := strings.Join(command, " ")
		if template, ok := DetectTemplateFromCommand(cmd); ok {
			fmt.Printf("No smd.toml found. Auto-detected '%s' from command.\n", template.Name)
			config, err = ConfigFromTemplate(template)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("No smd.toml found. Using dynamic mode (nixos + nix shell).")
			dynamicTemplate := templates["dynamic"]
			config, err = ConfigFromTemplate(&dynamicTemplate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}
			firstArg := command[0]
			os.Setenv("SMD_DYNAMIC_PACKAGE", firstArg)
		}
	}

	if tempMode {
		err = RunTempMode(config, command)
	} else {
		err = RunSessionMode(config, command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printInfo() {
	fmt.Printf(`smd v%s — Containerized development for AI agents

Every command runs in a persistent session container.
State (installed packages, files, config) survives across commands.

Usage:
  smd <command> [args]     Run in session container (persistent, default)
  smd -t <command> [args]  Run in temp container (ephemeral, --rm)
  smd --open               Interactive shell (live mode)
  smd --close              Destroy the session container

Setup:
  smd -i, --init           Create smd.toml with template selection
  smd --agentmd            Create AGENTS.md with smd usage instructions
  smd -h, --help           Full help and templates
  smd --version            Show version

Examples:
  smd apt install curl     Install packages persistently
  smd python script.py     Run in Python
  smd npm install           Install npm dependencies
  smd -t python test.py    Ephemeral — no state saved

Just run any command. State persists. No setup needed.
`, version)
}

func generateAgentMD() {
	content := "# Agent Instructions for this project\n\n" +
		"ALWAYS use `smd` for EVERY command. Never run commands directly.\n\n" +
		"## Why?\n" +
		"Every command runs in an isolated container with:\n" +
		"- **Persistent session** — apt, pip, npm, files all survive across commands\n" +
		"- **Automatic .env protection** — sensitive files are blocked in temp mode\n" +
		"- **Pre-configured runtimes** — no need to install Node, Python, Go, etc.\n\n" +
		"## All commands go through smd\n" +
		"```\n" +
		"smd ls                 # List files\n" +
		"smd cat file.txt       # Read files\n" +
		"smd head file.txt      # View file start\n" +
		"smd python script.py   # Run Python\n" +
		"smd npm install        # Install npm packages\n" +
		"smd go build           # Build Go project\n" +
		"smd apt install curl   # Install system packages\n" +
		"smd git status         # Git operations\n" +
		"smd pip install flask  # Install Python packages\n" +
		"smd make               # Run Make\n" +
		"smd cargo build        # Build Rust project\n" +
		"```\n\n" +
		"## Tips\n" +
		"- **State persists** — first `smd <cmd>` creates a session container, subsequent commands reuse it\n" +
		"- **Temp mode** — use `smd -t <cmd>` for a truly ephemeral one-shot command\n" +
		"- **Cleanup** — `smd --close` destroys the session container\n" +
		"- **First run** — just run `smd <cmd>` and everything auto-detects\n\n" +
		"This ensures all commands execute inside the project's containerized environment.\n"

	if err := os.WriteFile("AGENTS.md", []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating AGENTS.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created AGENTS.md with smd usage instructions.")
	fmt.Println("Edit AGENTS.md to add project-specific agent instructions.")
}

func printHelp() {
	fmt.Println(`smd - Secure My Directory

Usage:
  smd                          Show info page
  smd <command> [args]         Session mode (persistent container, default)
  smd -t <command> [args]      Temp mode (ephemeral container, --rm)
  smd -o, --open               Live mode (interactive shell)
  smd --close                  Destroy session container
  smd -i, --init               Create smd.toml with template selection
  smd --agentmd                Create AGENTS.md with smd usage instructions
  smd -h, --help               Show this help
  smd --version                Show version

Session mode (default):
  Creates a persistent container (<project>_s). All state survives.
  Container is created on first use, stopped between commands.

Temp mode (-t):
  Runs a command in an ephemeral container (<project>_t).
  Container is removed after the command. Fresh start every time.

Templates (use with --init):
  full              Everything installed
  dynamic           NixOS base - install on-demand via nix
  javascript/*      node, bun, all, demo
  python/*          python, pip, conda, all
  go, go/air        Go development
  rust              Rust with cargo
  elixir, gleam     Other languages

Examples:
  smd npm install          # Session: npm deps persist
  smd -t npm install        # Temp: fresh npm install each time
  smd --open                # Interactive shell
  smd python script.py      # Session: Python in persistent env
  smd --close               # Clean up session container

Configuration:
  smd.toml                  Auto-generated on first command

Environment Variables:
  SMD_IMAGE                 Override the container image
  SMD_DEBUG                 Show podman/docker commands being executed
  SMD_DYNAMIC_PACKAGE       Package for dynamic nix shell mode`)
}
