package main

import (
	"fmt"
	"os"
	"strings"
)

const version = "0.0.6"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("smd version %s\n", version)
		os.Exit(0)
	}

	// Handle --agentmd flag
	if len(os.Args) > 1 && os.Args[1] == "--agentmd" {
		generateAgentMD()
		os.Exit(0)
	}

	// Handle --init / -i flag
	if len(os.Args) > 1 && (os.Args[1] == "--init" || os.Args[1] == "-i") {
		_, err := TemplateInitInteractive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nRun 'smd --open' for live mode or 'smd <command>' for temp mode.\n")
		os.Exit(0)
	}

	// Show info page when no arguments
	if len(os.Args) == 1 {
		printInfo()
		os.Exit(0)
	}

	// Handle --open / -o flag: always enters live mode
	if os.Args[1] == "--open" || os.Args[1] == "-o" {
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
		os.Exit(0)
	}

	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// No config exists - try auto-detection from command
	if config == nil && len(os.Args) > 1 {
		cmd := strings.Join(os.Args[1:], " ")
		if template, ok := DetectTemplateFromCommand(cmd); ok {
			fmt.Printf("No smd.toml found. Auto-detected '%s' from command.\n", template.Name)
			config, err = ConfigFromTemplate(template)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Fallback: use dynamic mode (nixos)
			fmt.Println("No smd.toml found. Using dynamic mode (nixos + nix shell).")
			dynamicTemplate := templates["dynamic"]
			config, err = ConfigFromTemplate(&dynamicTemplate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}
			// Store the command as a nix package for dynamic execution
			firstArg := os.Args[1]
			os.Setenv("SMD_DYNAMIC_PACKAGE", firstArg)
		}
	}

	// Still no config - interactive init
	if config == nil {
		config, err = InteractiveInit()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
			os.Exit(1)
		}
	}

	args := os.Args[1:]
	err = RunTempMode(config, args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printInfo() {
	fmt.Printf(`smd v%s - Secure My Directory
Containerized development environments with automatic .env protection.

Usage:
  smd --open        Enter live mode (interactive shell with PWD mounted)
  smd <command>     Run command in temp/dynamic mode (ephemeral container)
  smd -i, --init    Initialize smd.toml with template selection
  smd --agentmd     Create/update AGENTS.md with smd usage instructions
  smd -h, --help    Show full help and templates
  smd --version     Show version
`, version)
}

func generateAgentMD() {
	content := `# Agent Instructions for this project

Always use \` + "`smd`" + ` for building and running commands instead of running tools directly.

- Instead of \` + "`go build`" + `, run \` + "`smd go build`" + `
- Instead of \` + "`npm install`" + `, run \` + "`smd npm install`" + `
- Instead of \` + "`python script.py`" + `, run \` + "`smd python script.py`" + `

This ensures all commands execute inside the project's containerized environment.
`

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
  smd                    Show info page
  smd -o, --open         Enter live mode (interactive shell with PWD mounted)
  smd <command> [args]   Run command in temp/dynamic mode (ephemeral container)
  smd -i, --init         Initialize smd.toml with template selection
  smd --agentmd          Create/update AGENTS.md with smd usage instructions
  smd -h, --help         Show this help
  smd --version          Show version

Templates (use with --init):
  full              Everything installed
  dynamic           NixOS base - install on-demand via nix
  javascript/*      node, bun, all, demo
  python/*          python, pip, conda, all
  go, go/air        Go development
  rust              Rust with cargo
  elixir, gleam     Other languages

Examples:
  smd --init             # Interactive template selection
  smd --open             # Live mode with smd.toml
  smd npm install        # Auto-detects node, creates temp config
  smd python script.py   # Auto-detects python, runs in isolated container
  smd go run .           # Dynamic mode if no config exists

Configuration:
  smd.toml               Project configuration (auto-generated if missing)

Environment Variables:
  SMD_IMAGE              Override container image
  SMD_DEBUG              Show podman commands being executed
  SMD_DYNAMIC_PACKAGE    Package for dynamic nix shell mode`)
}
