package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/leslierussell/tldg/internal/target"
)

// tldg-5xh

// Exit codes (spec §7.4).
const (
	ExitOK              = 0
	ExitGeneralFailure  = 1
	ExitBadArgs         = 2
	ExitTargetUnresolved = 3
	ExitAuthFailure     = 4
	ExitNetworkFailure  = 5
	ExitModelUnavailable = 6
	ExitPolicyDenied    = 7
	ExitIndexCorruption = 8
	ExitPartialResult   = 9
)

// errSilent signals a non-zero exit whose message was already rendered
// (e.g. the doctor report), so Execute must not print an extra error line.
var errSilent = errors.New("")

// codedError carries an explicit exit code.
type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string { return e.err.Error() }
func (e codedError) Unwrap() error { return e.err }

// coded wraps err with an explicit exit code.
func coded(code int, err error) error { return codedError{code: code, err: err} }

// Execute runs the root command and returns a process exit code.
func Execute() int {
	root := NewRootCmd()
	err := root.Execute()
	if err == nil {
		return ExitOK
	}
	if !errors.Is(err, errSilent) {
		fmt.Fprintln(os.Stderr, "tldg: "+err.Error())
	}
	return exitCode(err)
}

// classifyErr wraps a pipeline error with the most specific exit code available
// (spec §7.4). Already-coded errors pass through unchanged.
func classifyErr(err error) error {
	var ce codedError
	if errors.As(err, &ce) {
		return err
	}
	if errors.Is(err, target.ErrRemoteUnsupported) {
		return coded(ExitTargetUnresolved, err)
	}
	if errors.Is(err, os.ErrNotExist) {
		return coded(ExitTargetUnresolved, err)
	}
	return coded(ExitGeneralFailure, err)
}

// exitCode maps an error to a spec §7.4 exit code.
func exitCode(err error) int {
	var ce codedError
	if errors.As(err, &ce) {
		return ce.code
	}
	if errors.Is(err, target.ErrRemoteUnsupported) {
		return ExitTargetUnresolved
	}
	return ExitGeneralFailure
}
