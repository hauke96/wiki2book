package util

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
)

func Execute(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	var stderrBuffer bytes.Buffer
	errLog := io.MultiWriter(os.Stdout, &stderrBuffer)

	cmd.Stderr = errLog

	err := cmd.Run()

	if err != nil {
		return errors.Wrap(err, stderrBuffer.String())
	}

	return nil
}
