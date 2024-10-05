package main

import (
	"os"

	"nocomment/internal/cli"
	"nocomment/internal/langs"
	"nocomment/internal/process"
)

func main() {
	cfg, err := cli.Parse(os.Args[1:])
	if err != nil {
		os.Stderr.WriteString("nocomment: " + err.Error() + "\n")
		os.Exit(2)
	}
	processor := process.Processor{Registry: langs.Default(), Language: cfg.Lang}
	os.Exit(cli.Run(cfg, os.Stdin, os.Stdout, os.Stderr, processor))
}
