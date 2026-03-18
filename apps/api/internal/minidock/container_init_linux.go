//go:build linux

package minidock

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const containerInitCommand = "__minidock_container_init"

func HandleContainerInitCommand() bool {
	if len(os.Args) < 2 || os.Args[1] != containerInitCommand {
		return false
	}
	if err := runContainerInit(os.Args[2:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[container-linux] init error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
	return true
}

func runContainerInit(args []string) error {
	rootfs, hostname, command, commandArgs, err := parseContainerInitArgs(args)
	if err != nil {
		return err
	}
	if !isUsableRootfs(rootfs) {
		return fmt.Errorf("rootfs inválido: %s", rootfs)
	}
	if hostname == "" {
		hostname = "mdk-container"
	}

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("falha ao configurar hostname: %w", err)
	}
	if err := syscall.Mount("", "/", "", uintptr(syscall.MS_REC|syscall.MS_PRIVATE), ""); err != nil {
		return fmt.Errorf("falha ao isolar mount namespace: %w", err)
	}
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("falha em chroot: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("falha ao entrar no rootfs: %w", err)
	}

	if err := os.MkdirAll("/proc", 0o755); err != nil {
		return fmt.Errorf("falha ao preparar /proc: %w", err)
	}
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("falha ao montar /proc: %w", err)
	}
	_ = os.MkdirAll("/tmp", 0o777)

	resolved, err := resolveBinaryPath(command)
	if err != nil {
		return err
	}
	argv := append([]string{command}, commandArgs...)
	env := containerInitEnv()

	return syscall.Exec(resolved, argv, env)
}

func parseContainerInitArgs(args []string) (string, string, string, []string, error) {
	var rootfs string
	var hostname string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--rootfs":
			i++
			if i >= len(args) {
				return "", "", "", nil, errors.New("faltou valor para --rootfs")
			}
			rootfs = strings.TrimSpace(args[i])
		case "--hostname":
			i++
			if i >= len(args) {
				return "", "", "", nil, errors.New("faltou valor para --hostname")
			}
			hostname = strings.TrimSpace(args[i])
		case "--":
			cmd := args[i+1:]
			if len(cmd) == 0 {
				return "", "", "", nil, errors.New("comando ausente no init de container")
			}
			return rootfs, hostname, cmd[0], cloneStringSlice(cmd[1:]), nil
		default:
			return "", "", "", nil, fmt.Errorf("argumento desconhecido no init de container: %s", args[i])
		}
	}

	return "", "", "", nil, errors.New("faltou separador -- no init de container")
}

func resolveBinaryPath(command string) (string, error) {
	if strings.TrimSpace(command) == "" {
		return "", errors.New("comando vazio no init de container")
	}
	if strings.Contains(command, "/") {
		normalized := filepath.Clean(command)
		if fileExists(normalized) {
			return normalized, nil
		}
		return "", fmt.Errorf("binário não encontrado no rootfs: %s", command)
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("não foi possível resolver %s no rootfs: %w", command, err)
	}
	return path, nil
}

func containerInitEnv() []string {
	env := []string{
		"PATH=/bin:/usr/bin:/sbin:/usr/sbin",
		"HOME=/root",
		"TERM=xterm-256color",
	}
	if value := strings.TrimSpace(os.Getenv("TERM")); value != "" {
		env[2] = "TERM=" + value
	}
	return env
}
