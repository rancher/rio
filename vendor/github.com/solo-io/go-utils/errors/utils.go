package errors

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

func Wrapf(err error, format string, args ...interface{}) error {
	errString := fmt.Sprintf(format, args...)
	return xerrors.Errorf("%s: %w", errString, err)
}

func Errorf(format string, args ...interface{}) error {
	return xerrors.Errorf(format, args...)
}

func Errors(msgs []string) error {
	return xerrors.Errorf(strings.Join(msgs, "\n"))
}

func New(text string) error {
	return xerrors.New(text)
}

func Is(err, target error) bool {
	return xerrors.Is(err, target)
}
