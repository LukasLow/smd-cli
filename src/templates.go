package main

import (
	"fmt"
	"os"
	"strings"
)

type Template struct {
	Name        string
	Description string
	Image       string
	Type        []string
	Packages    []string // For nix dynamic mode
	IsNix       bool
}

var templates = map[string]Template{
	"full": {
		Name:        "full",
		Description: "Everything installed (Node, Python, Go, Rust, etc.)",
		Image:       "docker.io/nixos/nix:latest",
		Type:        []string{"full"},
		IsNix:       true,
	},
	"dynamic": {
		Name:        "dynamic",
		Description: "NixOS base - tools installed on-demand via nix shell",
		Image:       "docker.io/nixos/nix:latest",
		Type:        []string{"dynamic"},
		IsNix:       true,
	},
	"javascript/all": {
		Name:        "javascript/all",
		Description: "All JavaScript tools (Node + Bun + Deno)",
		Image:       "docker.io/nixos/nix:latest",
		Type:        []string{"javascript"},
		Packages:    []string{"nodejs", "bun", "deno"},
		IsNix:       true,
	},
	"javascript/node": {
		Name:        "javascript/node",
		Description: "Node.js only",
		Image:       "docker.io/node:20-alpine",
		Type:        []string{"nodejs"},
	},
	"javascript/bun": {
		Name:        "javascript/bun",
		Description: "Bun runtime",
		Image:       "docker.io/oven/bun:latest",
		Type:        []string{"bun"},
	},
	"javascript/demo": {
		Name:        "javascript/demo",
		Description: "Minimal JS demo environment",
		Image:       "docker.io/node:20-alpine",
		Type:        []string{"nodejs"},
	},
	"python": {
		Name:        "python",
		Description: "Python with pip",
		Image:       "docker.io/python:3.12-alpine",
		Type:        []string{"python"},
	},
	"python/all": {
		Name:        "python/all",
		Description: "Python with pip, poetry, conda",
		Image:       "docker.io/nixos/nix:latest",
		Type:        []string{"python"},
		Packages:    []string{"python3", "poetry", "conda"},
		IsNix:       true,
	},
	"python/pip": {
		Name:        "python/pip",
		Description: "Python with pip only",
		Image:       "docker.io/python:3.12-alpine",
		Type:        []string{"python"},
	},
	"python/conda": {
		Name:        "python/conda",
		Description: "Python with Conda",
		Image:       "docker.io/continuumio/miniconda3:latest",
		Type:        []string{"python"},
	},
	"go": {
		Name:        "go",
		Description: "Go standard",
		Image:       "docker.io/golang:1.22-alpine",
		Type:        []string{"go"},
	},
	"go/air": {
		Name:        "go/air",
		Description: "Go with Air live-reload",
		Image:       "docker.io/cosmtrek/air:latest",
		Type:        []string{"go"},
	},
	"rust": {
		Name:        "rust",
		Description: "Rust with cargo",
		Image:       "docker.io/rust:1.76-alpine",
		Type:        []string{"rust"},
	},
	"elixir": {
		Name:        "elixir",
		Description: "Elixir/Erlang",
		Image:       "docker.io/elixir:latest",
		Type:        []string{"elixir"},
	},
	"gleam": {
		Name:        "gleam",
		Description: "Gleam language",
		Image:       "docker.io/nixos/nix:latest",
		Type:        []string{"gleam"},
		Packages:    []string{"gleam"},
		IsNix:       true,
	},
}

var commandToTemplate = map[string]string{
	// JavaScript
	"npm":  "javascript/node",
	"npx":  "javascript/node",
	"node": "javascript/node",
	"yarn": "javascript/node",
	"pnpm": "javascript/node",
	"bun":  "javascript/bun",
	"deno": "javascript/all",

	// Python
	"python":  "python",
	"python3": "python",
	"pip":     "python/pip",
	"pip3":    "python/pip",
	"poetry":  "python/all",
	"conda":   "python/conda",

	// Go
	"go":    "go",
	"gofmt": "go",

	// Rust
	"cargo":  "rust",
	"rustc":  "rust",
	"rustup": "rust",

	// Elixir
	"elixir": "elixir",
	"mix":    "elixir",

	// Gleam
	"gleam": "gleam",
}

func GetTemplate(name string) (*Template, bool) {
	t, ok := templates[name]
	if !ok {
		return nil, false
	}
	return &t, true
}

func DetectTemplateFromCommand(command string) (*Template, bool) {
	// Extract first word (the actual command)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, false
	}

	cmd := parts[0]

	// Check for direct command mapping
	if templateName, ok := commandToTemplate[cmd]; ok {
		if t, ok := GetTemplate(templateName); ok {
			return t, true
		}
	}

	return nil, false
}

func ListTemplates() {
	fmt.Println("Available templates:")
	fmt.Println()

	categories := map[string][]Template{
		"General":    {templates["full"], templates["dynamic"]},
		"JavaScript": {templates["javascript/all"], templates["javascript/node"], templates["javascript/bun"], templates["javascript/demo"]},
		"Python":     {templates["python"], templates["python/all"], templates["python/pip"], templates["python/conda"]},
		"Go":         {templates["go"], templates["go/air"]},
		"Rust":       {templates["rust"]},
		"Elixir":     {templates["elixir"]},
		"Gleam":      {templates["gleam"]},
	}

	for cat, temps := range categories {
		fmt.Printf("  [%s]\n", cat)
		for _, t := range temps {
			fmt.Printf("    %-20s %s\n", t.Name, t.Description)
		}
		fmt.Println()
	}
}

func TemplateInitInteractive() (*Config, error) {
	fmt.Println("Initialize smd.toml with a template")
	fmt.Println()

	ListTemplates()

	fmt.Print("Select template: ")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	template, ok := GetTemplate(input)
	if !ok {
		return nil, fmt.Errorf("unknown template: %s", input)
	}

	return ConfigFromTemplate(template)
}

func ConfigFromTemplate(template *Template) (*Config, error) {
	projectName := GetProjectName()

	config := &Config{
		Project: ProjectConfig{
			Name: projectName,
			Type: template.Type,
		},
		Container: ContainerConfig{
			Image:   template.Image,
			Workdir: "/app",
		},
		Security: SecurityConfig{
			AllowEnv:   []string{},
			BlockFiles: defaultBlockFilesForTypes(template.Type),
		},
		Live: LiveConfig{
			MountPwd: true,
			Ports:    defaultPorts(template.Type),
		},
		Temp: TempConfig{
			AllowEnv: []string{},
			MountPwd: true,
		},
	}

	// Store nix packages in a special env var for dynamic mode
	if template.IsNix && len(template.Packages) > 0 {
		os.Setenv("SMD_NIX_PACKAGES", strings.Join(template.Packages, ","))
	}

	if err := SaveConfig(config); err != nil {
		return nil, err
	}

	fmt.Printf("Created smd.toml using template [%s]\n", template.Name)

	return config, nil
}

func defaultBlockFilesForTypes(types []string) []string {
	blocks := []string{}
	for _, t := range types {
		switch t {
		case "nodejs", "javascript":
			blocks = append(blocks, ".npmrc", ".yarnrc", ".pnpmfile.cjs")
		case "python":
			blocks = append(blocks, ".pypirc", "pip.conf")
		case "go":
			blocks = append(blocks, ".netrc")
		case "rust":
			blocks = append(blocks, ".cargo/credentials")
		}
	}
	return blocks
}

func IsDynamicMode(config *Config) bool {
	return config.Container.Image == "docker.io/nixos/nix:latest"
}

func GetNixPackages(config *Config) []string {
	if envPkgs := os.Getenv("SMD_NIX_PACKAGES"); envPkgs != "" {
		return strings.Split(envPkgs, ",")
	}
	return []string{}
}
