package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

var (
	workers int
	failed  []string
	lock    sync.Mutex
)

func init() {
	flag.IntVar(&workers, "workers", 10, "number of workers")
}

func main() {
	flag.Parse()

	if err := run(flag.Arg(0)); err != nil {
		log.Fatal(err)
	}
}

func run(path string) error {
	files := make(chan string)

	eg := errgroup.Group{}
	for i := 0; i < workers; i++ {
		eg.Go(func() error {
			return worker(files)
		})
	}

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".bats") {
			files <- path
		}
		return nil
	})
	if err != nil {
		return err
	}

	close(files)

	if err := eg.Wait(); err != nil {
		return err
	}

	for _, f := range failed {
		log.Println("FAILED", f)
	}

	if len(failed) > 0 {
		return errors.New("tests failed")
	}

	return nil
}

func worker(c chan string) error {
	for file := range c {
		log.Println("Running", file)
		cmd := exec.Command("bats", file)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			lock.Lock()
			failed = append(failed, file)
			lock.Unlock()
		}
	}

	return nil
}
