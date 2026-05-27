package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const configFile = "smd.toml"

type Config struct {
	Project   ProjectConfig   `toml:"project"`
	Container ContainerConfig `toml:"container"`
	Security  SecurityConfig  `toml:"security"`
	Volume    VolumeConfig    `toml:"volume"`
	Live      LiveConfig      `toml:"live"`
	Temp      TempConfig      `toml:"temp"`
}

type ProjectConfig struct {
	Name string   `toml:"name"`
	Type []string `toml:"type"`
}

type ContainerConfig struct {
	Image   string `toml:"image"`
	Workdir string `toml:"workdir"`
	Runtime string `toml:"runtime"`
}

type SecurityConfig struct {
	AllowEnv   []string `toml:"allow_env"`
	BlockFiles []string `toml:"block_files"`
}

type LiveConfig struct {
	MountPwd bool     `toml:"mount_pwd"`
	Ports    []string `toml:"ports"`
}

type TempConfig struct {
	AllowEnv []string `toml:"allow_env"`
	MountPwd bool     `toml:"mount_pwd"`
	ReadOnly bool     `toml:"read_only"`
}

type VolumeConfig struct {
	Enabled  bool            `toml:"enabled"`
	Packages map[string]bool `toml:"packages"`
}

type PackageCache struct {
	VolumeName string
	MountPath  string
}

var PackageCachePaths = map[string]PackageCache{
	"npm":   {VolumeName: "smd_pkg_npm",   MountPath: "/root/.npm"},
	"yarn":  {VolumeName: "smd_pkg_yarn",  MountPath: "/root/.yarn"},
	"pnpm":  {VolumeName: "smd_pkg_pnpm",  MountPath: "/root/.local/share/pnpm"},
	"pip":   {VolumeName: "smd_pkg_pip",   MountPath: "/root/.cache/pip"},
	"poetry": {VolumeName: "smd_pkg_poetry", MountPath: "/root/.cache/pypoetry"},
	"go":    {VolumeName: "smd_pkg_go",    MountPath: "/go/pkg/mod"},
	"cargo": {VolumeName: "smd_pkg_cargo", MountPath: "/usr/local/cargo/registry"},
	"bun":   {VolumeName: "smd_pkg_bun",   MountPath: "/root/.bun/install/cache"},
	"conda": {VolumeName: "smd_pkg_conda", MountPath: "/opt/conda/pkgs"},
}

func defaultVolumesForTypes(types []string) map[string]bool {
	volumes := map[string]bool{}
	for _, t := range types {
		switch t {
		case "nodejs", "javascript":
			volumes["npm"] = true
		case "python":
			volumes["pip"] = true
		case "go":
			volumes["go"] = true
		case "rust":
			volumes["cargo"] = true
		case "bun":
			volumes["bun"] = true
		case "conda":
			volumes["conda"] = true
		}
	}
	// Always enable if any volumes were detected
	if len(volumes) > 0 {
		return volumes
	}
	return nil
}

var supportedEnvironments = map[string]string{
	"nodejs": "docker.io/node:20-alpine",
	"python": "docker.io/python:3.12-alpine",
	"go":     "docker.io/golang:1.22-alpine",
	"rust":   "docker.io/rust:1.76-alpine",
}

func LoadConfig() (*Config, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, nil
	}

	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configFile, err)
	}

	if config.Container.Workdir == "" {
		config.Container.Workdir = "/app"
	}

	return &config, nil
}

func InteractiveInit() (*Config, error) {
	fmt.Println("No smd.toml found. Let's create one!")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Available environments:")
	fmt.Println("  1) nodejs - Node.js applications")
	fmt.Println("  2) python - Python applications")
	fmt.Println("  3) go     - Go applications")
	fmt.Println("  4) rust   - Rust applications")
	fmt.Println()
	fmt.Print("Select environment(s) (comma-separated numbers, e.g., 1,3): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	selectedTypes := parseSelection(strings.TrimSpace(input))
	if len(selectedTypes) == 0 {
		return nil, fmt.Errorf("no environment selected")
	}

	fmt.Println()
	fmt.Println("Choose mode:")
	fmt.Println("  [temp] - Secure/isolated (one-off commands, no file sync)")
	fmt.Println("  [live] - Persistent/mounted (development with file watching)")
	fmt.Print("Select mode [temp/live]: ")

	modeInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	mode := strings.TrimSpace(strings.ToLower(modeInput))
	if mode != "live" {
		mode = "temp"
	}

	fmt.Println()
	fmt.Print("Project name [my-project]: ")
	nameInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	projectName := strings.TrimSpace(nameInput)
	if projectName == "" {
		projectName = "my-project"
	}

	volumes := defaultVolumesForTypes(selectedTypes)
	config := &Config{
		Project: ProjectConfig{
			Name: projectName,
			Type: selectedTypes,
		},
		Container: ContainerConfig{
			Image:   supportedEnvironments[selectedTypes[0]],
			Workdir: "/app",
		},
		Security: SecurityConfig{
			AllowEnv:   []string{},
			BlockFiles: defaultBlockFilesForTypes(selectedTypes),
		},
		Volume: VolumeConfig{
			Enabled:  len(volumes) > 0,
			Packages: volumes,
		},
		Live: LiveConfig{
			MountPwd: true,
			Ports:    defaultPorts(selectedTypes),
		},
		Temp: TempConfig{
			AllowEnv: []string{},
			MountPwd: true,
		},
	}

	if err := SaveConfig(config); err != nil {
		return nil, err
	}

	modeHint := "'smd --open' for live mode | 'smd <command>' for temp mode"
	if mode == "live" {
		modeHint = "'smd --open' to start live mode"
	} else {
		modeHint = "'smd <command>' to execute commands in temp mode"
	}
	fmt.Printf("\nCreated %s with %s mode preference.\n", configFile, mode)
	fmt.Printf("Run %s.\n", modeHint)

	return config, nil
}

func parseSelection(input string) []string {
	var result []string
	parts := strings.Split(input, ",")

	for _, part := range parts {
		n := strings.TrimSpace(part)
		switch n {
		case "1":
			result = append(result, "nodejs")
		case "2":
			result = append(result, "python")
		case "3":
			result = append(result, "go")
		case "4":
			result = append(result, "rust")
		}
	}

	return result
}

func typePorts(t string) []string {
	switch t {
	case "nodejs":
		return []string{"3000:3000", "8080:8080"}
	case "python":
		return []string{"5000:5000", "8000:8000"}
	case "go":
		return []string{"8080:8080"}
	case "rust":
		return []string{"8080:8080"}
	default:
		return nil
	}
}

func defaultPorts(types []string) []string {
	seen := map[string]bool{}
	var ports []string
	for _, t := range types {
		for _, p := range typePorts(t) {
			if !seen[p] {
				seen[p] = true
				ports = append(ports, p)
			}
		}
	}
	sort.Strings(ports)
	return ports
}

func SaveConfig(config *Config) error {
	tmpFile := configFile + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", tmpFile, err)
	}

	if err := toml.NewEncoder(f).Encode(config); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("failed to encode config: %w", err)
	}
	f.Close()

	if err := os.Rename(tmpFile, configFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func FindEnvFiles() ([]string, error) {
	var envFiles []string

	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".env") {
			envFiles = append(envFiles, name)
		}
	}

	return envFiles, nil
}

func isAllowedEnvFile(filename string, config *Config) bool {
	for _, allowed := range config.Security.AllowEnv {
		if allowed == filename {
			return true
		}
	}
	return false
}

func IsEnvFileBlocked(filename string, config *Config, isTempMode bool) bool {
	if !isTempMode {
		return false
	}

	if !strings.HasPrefix(filename, ".env") {
		return false
	}

	return !isAllowedEnvFile(filename, config)
}

func GetProjectName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "smd-project"
	}
	return filepath.Base(cwd)
}
