package util

import (
	"github.com/hauke96/sigolo/v2"
	"github.com/pkg/errors"
)

/*
Contracts are one way to ensure, that the programmer did his or her job by passing correct data and catching faulty data
before passing it to a function. Therefore, contracts are no usual "something's wrong, I'll return an error" but a
stricter "oh no, the programmer screwed up, we have to stop NOW". Therefore, there is not way to recover from a contract
violation since we intentionally panic here.
*/

func Require(condition bool) {
	if !condition {
		sigolo.Fatalf("%+v", errors.New("Contract violation! This might indicate a bug in the software."))
	}
}

func Requirem(condition bool, message string) {
	if !condition {
		sigolo.Fatalf("%+v", errors.New("Contract violation! This might indicate a bug in the software. Details: "+message))
	}
}

func Requiref(condition bool, format string, args ...interface{}) {
	if !condition {
		sigolo.Fatalf("%+v", errors.Errorf("Contract violation! This might indicate a bug in the software. Details: "+format, args...))
	}
}
