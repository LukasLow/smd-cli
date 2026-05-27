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
		"--name", fmt.Sprintf("%s-live", config.Project.Name),
	}

	args = append(args, buildVolumeArgs(config)...)

	for _, port := range config.Live.Ports {
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

func buildVolumeArgs(config *Config) []string {
	var args []string
	if config == nil || !config.Volume.Enabled || len(config.Volume.Packages) == 0 {
		return args
	}

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
	return args
}
