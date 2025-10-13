package terminal
import (
    "bytes"
    "os"
    "os/exec"
    "path/filepath"
)

func RunSyncCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}else{
		return nil
	}
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

// RunOutput runs a command and returns its stdout as a trimmed string.
func RunOutput(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
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
