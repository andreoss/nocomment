package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"
)

const usageText = `usage: nocomment [flags] [path ...]

nocomment removes comments from source files.

flags:
  -w           write result to source file instead of standard output
  -d           display diffs instead of rewriting files
  -l           list files whose result differs
  -e           report all errors, not only the first
  -lang name   force the language for paths and standard input
  -version     print version and build metadata
  -h, -help    show this help
`

var Version = "devel"

type Config struct {
	Write     bool
	Diff      bool
	List      bool
	AllErrors bool
	Lang      string
	Help      bool
	Version   bool
	Paths     []string
}

type Processor interface {
	Process(path string, src []byte) ([]byte, error)
}

func Parse(args []string) (Config, error) {
	var cfg Config
	fs := flag.NewFlagSet("nocomment", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&cfg.Write, "w", false, "")
	fs.BoolVar(&cfg.Diff, "d", false, "")
	fs.BoolVar(&cfg.List, "l", false, "")
	fs.BoolVar(&cfg.AllErrors, "e", false, "")
	fs.StringVar(&cfg.Lang, "lang", "", "")
	fs.BoolVar(&cfg.Help, "h", false, "")
	fs.BoolVar(&cfg.Help, "help", false, "")
	fs.BoolVar(&cfg.Version, "version", false, "")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	cfg.Paths = fs.Args()
	return cfg, nil
}

func Run(cfg Config, stdin io.Reader, stdout, stderr io.Writer, processor Processor) int {
	return runWith(cfg, stdin, stdout, stderr, processor, osFS{})
}

func runWith(cfg Config, stdin io.Reader, stdout, stderr io.Writer, processor Processor, files FileSystem) int {
	if cfg.Help {
		fmt.Fprint(stdout, usageText)
		return 0
	}
	if cfg.Version {
		fmt.Fprintln(stdout, versionString())
		return 0
	}
	if cfg.Write && (cfg.Diff || cfg.List) {
		fmt.Fprintln(stderr, "nocomment: cannot combine -w with -d or -l")
		return 2
	}
	if cfg.Diff && cfg.List {
		fmt.Fprintln(stderr, "nocomment: cannot combine -d with -l")
		return 2
	}
	if len(cfg.Paths) == 0 {
		return runStdin(cfg, stdin, stdout, stderr, processor)
	}
	status := 0
	for _, path := range cfg.Paths {
		if err := processFile(cfg, path, stdout, processor, files); err != nil {
			fmt.Fprintf(stderr, "nocomment: %v\n", err)
			status = 2
			if !cfg.AllErrors {
				return status
			}
		}
	}
	return status
}

func runStdin(cfg Config, stdin io.Reader, stdout, stderr io.Writer, processor Processor) int {
	if cfg.Write {
		fmt.Fprintln(stderr, "nocomment: cannot use -w with standard input")
		return 2
	}
	src, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "nocomment: %v\n", err)
		return 2
	}
	out, err := processor.Process("", src)
	if err != nil {
		fmt.Fprintf(stderr, "nocomment: %v\n", err)
		return 2
	}
	if bytes.Equal(src, out) {
		if !cfg.List && !cfg.Diff {
			if _, err := stdout.Write(out); err != nil {
				fmt.Fprintf(stderr, "nocomment: %v\n", err)
				return 2
			}
		}
		return 0
	}
	switch {
	case cfg.List:
		fmt.Fprintln(stdout, "<standard input>")
	case cfg.Diff:
		if _, err := stdout.Write(UnifiedDiff("<standard input>", src, out)); err != nil {
			fmt.Fprintf(stderr, "nocomment: %v\n", err)
			return 2
		}
	default:
		if _, err := stdout.Write(out); err != nil {
			fmt.Fprintf(stderr, "nocomment: %v\n", err)
			return 2
		}
	}
	return 0
}

func processFile(cfg Config, path string, stdout io.Writer, processor Processor, files FileSystem) error {
	info, err := files.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s: is a directory", path)
	}
	src, err := files.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := processor.Process(path, src)
	if err != nil {
		return err
	}
	changed := !bytes.Equal(src, out)
	switch {
	case cfg.Write:
		if !changed {
			return nil
		}
		return files.WriteFile(path, out, info.Mode())
	case cfg.List:
		if changed {
			if _, err := fmt.Fprintln(stdout, path); err != nil {
				return err
			}
		}
		return nil
	case cfg.Diff:
		if changed {
			if _, err := stdout.Write(UnifiedDiff(path, src, out)); err != nil {
				return err
			}
		}
		return nil
	default:
		_, err = stdout.Write(out)
		return err
	}
}

func versionString() string {
	var b strings.Builder
	b.WriteString("nocomment ")
	b.WriteString(Version)
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			b.WriteString(" ")
			b.WriteString(v)
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision", "vcs.time":
				if s.Value != "" {
					b.WriteString(" ")
					b.WriteString(s.Value)
				}
			}
		}
	}
	b.WriteString(" ")
	b.WriteString(runtime.Version())
	return b.String()
}
