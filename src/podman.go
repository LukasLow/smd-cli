package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunLiveMode(config *Config) error {
	args := []string{
		"run",
		"--rm",
		"-it",
		"-v", fmt.Sprintf("%s:%s", getPwd(), config.Container.Workdir),
		"-w", config.Container.Workdir,
		"--name", fmt.Sprintf("%s-live", config.Project.Name),
	}

	for _, port := range config.Live.Ports {
		args = append(args, "-p", port)
	}

	envFiles, err := FindEnvFiles()
	if err != nil {
		return fmt.Errorf("failed to scan for env files: %w", err)
	}

	blockedFiles := checkBlockedFiles(envFiles, config, false)
	if len(blockedFiles) > 0 {
		fmt.Println("Note: The following sensitive files are available in live mode:")
		for _, f := range blockedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}

	for _, envFile := range envFiles {
		args = append(args, "--env-file", envFile)
	}

	image := getImage(config)
	args = append(args, image)

	if os.Getenv("SMD_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "podman %s\n", strings.Join(args, " "))
	}

	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunTempMode(config *Config, command []string) error {
	args := []string{
		"run",
		"--rm",
	}

	envFiles, err := FindEnvFiles()
	if err != nil {
		return fmt.Errorf("failed to scan for env files: %w", err)
	}

	blockedFiles := checkBlockedFiles(envFiles, config, true)
	if len(blockedFiles) > 0 {
		fmt.Println("Safety Shield: Blocking sensitive files in temp mode:")
		for _, f := range blockedFiles {
			fmt.Printf("  - %s (blocked)\n", f)
		}
	}

	for _, envFile := range envFiles {
		if !IsEnvFileBlocked(envFile, config, true) {
			args = append(args, "--env-file", envFile)
		}
	}

	// Mount PWD by default in temp mode (can be disabled in config)
	if config.Temp.MountPwd {
		mountOpt := ""
		if config.Temp.ReadOnly {
			mountOpt = ":ro"
		}
		args = append(args, "-v", fmt.Sprintf("%s:%s%s", getPwd(), config.Container.Workdir, mountOpt))
	}

	args = append(args, "-w", config.Container.Workdir)

	image := getImage(config)
	args = append(args, image)

	// Check for dynamic nix mode
	if IsDynamicMode(config) {
		// Handle dynamic nix shell execution
		nixPackages := GetNixPackages(config)
		if dynamicPkg := os.Getenv("SMD_DYNAMIC_PACKAGE"); dynamicPkg != "" {
			nixPackages = append([]string{dynamicPkg}, nixPackages...)
		}

		if len(nixPackages) > 0 {
			// Build nix shell command: nix shell nixpkgs#pkg1 nixpkgs#pkg2 -- command
			nixArgs := []string{"shell"}
			for _, pkg := range nixPackages {
				nixArgs = append(nixArgs, "nixpkgs#"+pkg)
			}
			nixArgs = append(nixArgs, "--")
			nixArgs = append(nixArgs, command...)

			args = append(args, "-c", "nix", "run", "nixpkgs#nix", "--", "shell")
			for _, pkg := range nixPackages {
				args = append(args, "nixpkgs#"+pkg)
			}
			args = append(args, "--")
			args = append(args, command...)
		} else {
			args = append(args, command...)
		}
	} else {
		args = append(args, command...)
	}

	if os.Getenv("SMD_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "podman %s\n", strings.Join(args, " "))
	}

	cmd := exec.Command("podman", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func checkBlockedFiles(envFiles []string, config *Config, isTempMode bool) []string {
	var blocked []string
	for _, f := range envFiles {
		if IsEnvFileBlocked(f, config, isTempMode) {
			blocked = append(blocked, f)
		}
	}
	return blocked
}

func getImage(config *Config) string {
	if envImage := os.Getenv("SMD_IMAGE"); envImage != "" {
		return envImage
	}
	return config.Container.Image
}

func getPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return pwd
}
