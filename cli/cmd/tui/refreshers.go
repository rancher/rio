package tui

import (
	"bytes"
	"os"
	"strings"

	"github.com/docker/docker/pkg/reexec"
	"github.com/pkg/errors"
)

/*
	Refresher refreshes the data by invoking the defined functions. Right now refreshers are invoked by shell output,
	but it can customized by override buffer writer.
*/
var (
	ConfigRefresher = func(b *bytes.Buffer) error {
		args := []string{"config"}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	PublicDomainRefresher = func(b *bytes.Buffer) error {
		args := []string{"domain"}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	ExternalRefresher = func(b *bytes.Buffer) error {
		args := []string{"external"}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	RouteRefresher = func(b *bytes.Buffer) error {
		args := []string{"route"}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	AppRefresher = func(b *bytes.Buffer) error {
		args := []string{"ps"}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	ServiceRefresher = func(b *bytes.Buffer) error {
		args := []string{"revision"}
		if *servicePrefix != "" {
			args = append(args, *servicePrefix)
		}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	PodRefresher = func(b *bytes.Buffer) error {
		args := []string{"ps", "-c"}
		if *podPrefix != "" {
			args = append(args, *podPrefix)
		}
		if showSystem {
			args = append([]string{"--system"}, args...)
		}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}

	BuildRefresher = func(b *bytes.Buffer) error {
		args := []string{"build"}
		cmd := reexec.Command(append([]string{"rio"}, args...)...)
		errBuffer := &strings.Builder{}
		cmd.Env = append(os.Environ(), "FORMAT=raw")
		cmd.Stdout = b
		cmd.Stderr = errBuffer
		if err := cmd.Run(); err != nil {
			return errors.New(errBuffer.String())
		}
		return nil
	}
)
