package main

import (
	"errors"
	"os"

	"nocomment/internal/cli"
)

type bootstrap struct{}

func (bootstrap) Process(path string, src []byte) ([]byte, error) {
	return nil, errors.New("no language is registered")
}

func main() {
	cfg, err := cli.Parse(os.Args[1:])
	if err != nil {
		os.Stderr.WriteString("nocomment: " + err.Error() + "\n")
		os.Exit(2)
	}
	os.Exit(cli.Run(cfg, os.Stdin, os.Stdout, os.Stderr, bootstrap{}))
}
