package util

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

func ExecuteCommandWithArgs(commandString string, workingDirectory string) error {
	commandParts := strings.Split(commandString, " ")
	commandExecutable := commandParts[0]
	commandArgs := commandParts[1:]

	return Execute(commandExecutable, workingDirectory, commandArgs...)
}

func Execute(name string, workingDirectory string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	var stderrBuffer strings.Builder
	cmd.Stderr = &stderrBuffer
	cmd.Dir = workingDirectory

	sigolo.Debugf("Execute command: %s", cmd.String())
	err := cmd.Run()

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Command \"%s\" exited with error: %s", cmd.String(), stderrBuffer.String()))
	}

	return nil
}
