package exec

import (
	"bytes"
	"os"
	osexec "os/exec"
	"path/filepath"
)

func RunSyncCommand(name string, args ...string) error {
	cmd := osexec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func RunAsyncCommand(name string, args ...string) (*osexec.Cmd, error) {
	cmd := osexec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func RunOutput(name string, args ...string) (string, error) {
	cmd := osexec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out.Bytes())), nil
}

func RunSyncUserShell(cmdline string) error {
	cmd := userShellCommandLinux(cmdline)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunAsyncUserShell(cmdline string) (*osexec.Cmd, error) {
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

func userShellCommandLinux(cmdline string) *osexec.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	name := filepath.Base(shell)

	switch name {
	case "bash":
		return osexec.Command(shell, "-lc", cmdline)
	case "zsh":
		return osexec.Command(shell, "-lc", cmdline)
	case "fish":
		return osexec.Command(shell, "-lc", cmdline)
	case "dash", "sh":
		return osexec.Command(shell, "-c", cmdline)
	default:
		return osexec.Command(shell, "-c", cmdline)
	}
}
