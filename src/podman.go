package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func RunLiveMode(config *Config) error {
	args := []string{
		"run",
		"--rm",
		"-it",
		"-v", fmt.Sprintf("%s:%s", getPwd(), config.Container.Workdir),
		"-w", config.Container.Workdir,
		"--name", liveContainerName(config),
	}

	args = append(args, buildVolumeArgs(config)...)

	for _, port := range getPorts(config.Live.Ports, config) {
		args = append(args, "-p", port)
	}

	envFiles, err := FindEnvFiles()
	if err != nil {
		return fmt.Errorf("failed to scan for env files: %w", err)
	}

	var unlisted []string
	for _, envFile := range envFiles {
		if !isAllowedEnvFile(envFile, config) {
			unlisted = append(unlisted, envFile)
		}
	}
	if len(unlisted) > 0 {
		fmt.Println("Live mode: the following .env files are mounted (not in allow list):")
		for _, f := range unlisted {
			fmt.Printf("  - %s\n", f)
		}
	}

	for _, envFile := range envFiles {
		args = append(args, "--env-file", envFile)
	}

	image := getImage(config)
	args = append(args, image)

	runtime := runtimeName(config)
	if os.Getenv("SMD_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", runtime, strings.Join(args, " "))
	}

	cmd := exec.Command(runtime, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunTempMode(config *Config, command []string) error {
	args := []string{
		"run",
		"--rm",
		"--name", containerName(config, "t"),
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

	args = append(args, buildVolumeArgs(config)...)

	for _, port := range getPorts(nil, config) {
		args = append(args, "-p", port)
	}

	args = append(args, "-w", config.Container.Workdir)

	image := getImage(config)
	args = append(args, image)

	if IsDynamicMode(config) {
		nixPackages := GetNixPackages(config)
		if dynamicPkg := os.Getenv("SMD_DYNAMIC_PACKAGE"); dynamicPkg != "" {
			nixPackages = append([]string{dynamicPkg}, nixPackages...)
		}

		if len(nixPackages) > 0 {
			args = append(args, "shell")
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

	runtime := runtimeName(config)
	if os.Getenv("SMD_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", runtime, strings.Join(args, " "))
	}

	cmd := exec.Command(runtime, args...)
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

func runtimeName(config *Config) string {
	if config != nil && config.Container.Runtime != "" {
		return config.Container.Runtime
	}
	return "docker"
}

func RunSessionMode(config *Config, command []string) error {
	name := containerName(config, "s")
	runtime := runtimeName(config)

	if err := ensureSessionContainer(config, name, runtime); err != nil {
		return err
	}

	if err := exec.Command(runtime, "start", name).Run(); err != nil {
		return fmt.Errorf("failed to start session container: %w", err)
	}
	defer exec.Command(runtime, "stop", name).Run()

	execArgs := []string{"exec", "-i", "-w", config.Container.Workdir, name}

	if IsDynamicMode(config) {
		nixPkgs := GetNixPackages(config)
		if len(nixPkgs) > 0 {
			execArgs = append(execArgs, "nix", "shell")
			for _, p := range nixPkgs {
				execArgs = append(execArgs, "nixpkgs#"+p)
			}
			execArgs = append(execArgs, "--")
		}
	}

	execArgs = append(execArgs, command...)

	if os.Getenv("SMD_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", runtime, strings.Join(execArgs, " "))
	}

	cmd := exec.Command(runtime, execArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ensureSessionContainer(config *Config, name, runtime string) error {
	check := exec.Command(runtime, "ps", "-a", "--filter", fmt.Sprintf("name=^%s$", name), "--format", "{{.Names}}")
	out, _ := check.Output()
	if strings.TrimSpace(string(out)) != "" {
		return nil
	}

	createArgs := []string{
		"create",
		"--name", name,
		"-v", fmt.Sprintf("%s:%s", getPwd(), config.Container.Workdir),
		"-w", config.Container.Workdir,
	}

	createArgs = append(createArgs, buildVolumeArgs(config)...)

	for _, port := range getPorts(config.Session.Ports, config) {
		createArgs = append(createArgs, "-p", port)
	}

	envFiles, _ := FindEnvFiles()
	var unlisted []string
	for _, ef := range envFiles {
		if !isAllowedEnvFile(ef, config) {
			unlisted = append(unlisted, ef)
		}
	}
	if len(unlisted) > 0 {
		fmt.Println("Session container created with .env files (not in whitelist):")
		for _, f := range unlisted {
			fmt.Printf("  - %s\n", f)
		}
	}
	for _, ef := range envFiles {
		createArgs = append(createArgs, "--env-file", ef)
	}

	createArgs = append(createArgs, getImage(config), "sleep", "infinity")

	if os.Getenv("SMD_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", runtime, strings.Join(createArgs, " "))
	}

	if err := exec.Command(runtime, createArgs...).Run(); err != nil {
		return fmt.Errorf("failed to create session container: %w", err)
	}
	fmt.Printf("Created session container: %s\n", name)
	return nil
}

func DestroySessionContainer(config *Config) error {
	name := containerName(config, "s")
	runtime := runtimeName(config)

	exec.Command(runtime, "stop", name).Run()
	err := exec.Command(runtime, "rm", "-f", name).Run()
	if err != nil {
		return fmt.Errorf("failed to remove session container: %w", err)
	}
	fmt.Printf("Destroyed session container: %s\n", name)
	return nil
}

func buildVolumeArgs(config *Config) []string {
	var args []string
	if config == nil || !config.Volume.Enabled {
		return args
	}

	// Package cache volumes
	var names []string
	for pkgName, enabled := range config.Volume.Packages {
		if enabled {
			if _, ok := PackageCachePaths[pkgName]; ok {
				names = append(names, pkgName)
			}
		}
	}
	sort.Strings(names)

	for _, name := range names {
		cache := PackageCachePaths[name]
		args = append(args, "-v", fmt.Sprintf("%s:%s", cache.VolumeName, cache.MountPath))
	}

	// User-defined volume mounts (named volumes or bind mounts)
	for _, m := range config.Volume.Mounts {
		args = append(args, "-v", m)
	}

	return args
}
