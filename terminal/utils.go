package terminal

import (
	"os"
	"os/exec"
	"path/filepath"
)

func RunSyncCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunAsyncCommand(name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

func RunSyncUserShell(cmdline string) error {
	cmd := userShellCommandLinux(cmdline)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunAsyncUserShell(cmdline string) (*exec.Cmd, error) {
	cmd := userShellCommandLinux(cmdline)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func userShellCommandLinux(cmdline string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	name := filepath.Base(shell)

	switch name {
	case "bash":
		return exec.Command(shell, "-lc", cmdline)
	case "zsh":
		return exec.Command(shell, "-lc", cmdline)
	case "fish":
		return exec.Command(shell, "-lc", cmdline)
	case "dash", "sh":
		return exec.Command(shell, "-c", cmdline)
	default:
		return exec.Command(shell, "-c", cmdline)
	}
}
