package util

import (
	"bytes"
	"fmt"
	"os/exec"
)

func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %s %v\nError: %w\nStderr: %s", name, args, err, stderr.String())
	}

	return stdout.String(), nil
}
