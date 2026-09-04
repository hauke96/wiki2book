package util

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

func ExecuteCommandWithArgs(commandString string, workingDirectory string) error {
	commandParts := []string{}

	insideString := false
	tmpCommand := ""
	for _, char := range commandString {
		if char == ' ' {
			if !insideString {
				commandParts = append(commandParts, tmpCommand)
				tmpCommand = ""
				continue
			}
		} else if char == '"' {
			insideString = !insideString // flip the flag
			continue
		}

		tmpCommand += string(char)
	}
	commandParts = append(commandParts, tmpCommand) // adding the last part of the command string

	commandExecutable := commandParts[0]
	commandArgs := commandParts[1:]

	return Execute(commandExecutable, workingDirectory, commandArgs...)
}

func Execute(name string, workingDirectory string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	var stderrBuffer strings.Builder
	cmd.Stderr = &stderrBuffer

	var stdoutBuffer strings.Builder
	cmd.Stdout = &stdoutBuffer

	cmd.Dir = workingDirectory

	sigolo.Debugf("Execute command: %s", cmd.String())
	err := cmd.Run()

	sigolo.Tracef("Output:\n%s", stdoutBuffer.String())

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Command \"%s\" exited with error: %s", cmd.String(), stderrBuffer.String()))
	}

	return nil
}
