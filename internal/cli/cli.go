package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
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
  -s pattern   process only files whose name matches a comma-separated glob
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
	Select    string
	Help      bool
	Version   bool
	Paths     []string
}

type Processor interface {
	Process(path string, src []byte) ([]byte, error)
	Supports(path string) bool
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
	fs.StringVar(&cfg.Select, "s", "", "")
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
	sel, err := newSelector(cfg.Select)
	if err != nil {
		fmt.Fprintf(stderr, "nocomment: %v\n", err)
		return 2
	}
	status := 0
	for _, path := range cfg.Paths {
		targets, err := expandPath(path, sel, processor, files)
		if err != nil {
			fmt.Fprintf(stderr, "nocomment: %v\n", err)
			status = 2
			if !cfg.AllErrors {
				return status
			}
			continue
		}
		for _, target := range targets {
			if err := processFile(cfg, target, stdout, processor, files); err != nil {
				fmt.Fprintf(stderr, "nocomment: %v\n", err)
				status = 2
				if !cfg.AllErrors {
					return status
				}
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
	path := ""
	if cfg.Lang == "" {
		ext := shebangExtension(src)
		if ext == "" {
			fmt.Fprintln(stderr, "nocomment: standard input requires -lang or a recognized shebang")
			return 2
		}
		path = "stdin" + ext
	}
	out, err := processor.Process(path, src)
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
		return fmt.Errorf("%s: %w", path, err)
	}
	changed := !bytes.Equal(src, out)
	switch {
	case cfg.Write:
		if !changed {
			return nil
		}
		if info.Mode().Perm()&0o200 == 0 {
			return fmt.Errorf("%s: refusing to write read-only file", path)
		}
		link, err := files.Lstat(path)
		if err != nil {
			return err
		}
		if link.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s: refusing to write through symlink", path)
		}
		return writeAtomic(files, path, out, info.Mode())
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

type selector []string

func newSelector(spec string) (selector, error) {
	if spec == "" {
		return nil, nil
	}
	var out selector
	for _, pattern := range strings.Split(spec, ",") {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if _, err := filepath.Match(pattern, "name"); err != nil {
			return nil, fmt.Errorf("-s pattern %q: %v", pattern, err)
		}
		out = append(out, pattern)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("-s needs a pattern")
	}
	return out, nil
}

func (s selector) selects(name string) bool {
	if len(s) == 0 {
		return true
	}
	for _, pattern := range s {
		if ok, err := filepath.Match(pattern, name); ok && err == nil {
			return true
		}
	}
	return false
}

func expandPath(path string, sel selector, processor Processor, files FileSystem) ([]string, error) {
	info, err := files.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	return walkDir(path, sel, processor, files)
}

func walkDir(root string, sel selector, processor Processor, files FileSystem) ([]string, error) {
	entries, err := files.ReadDir(root)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var out []string
	for _, entry := range entries {
		name := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			if skipDir(entry.Name()) {
				continue
			}
			sub, err := walkDir(name, sel, processor, files)
			if err != nil {
				return nil, err
			}
			out = append(out, sub...)
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !sel.selects(entry.Name()) {
			continue
		}
		if !processor.Supports(name) {
			continue
		}
		out = append(out, name)
	}
	return out, nil
}

func skipDir(name string) bool {
	if name == "" || strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "vendor", "node_modules", "bin", "build", "dist", "target", "scratch", "_build":
		return true
	}
	return false
}

func shebangExtension(src []byte) string {
	if !bytes.HasPrefix(src, []byte("#!")) {
		return ""
	}
	line := src
	if i := bytes.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	fields := strings.Fields(string(line))
	if len(fields) == 0 {
		return ""
	}
	interp := filepath.Base(fields[0])
	if interp == "env" {
		interp = ""
		for _, f := range fields[1:] {
			if strings.HasPrefix(f, "-") {
				continue
			}
			interp = filepath.Base(f)
			break
		}
	}
	switch trimVersion(interp) {
	case "python":
		return ".py"
	case "node", "nodejs":
		return ".js"
	case "sh", "bash", "dash", "zsh", "ksh":
		return ".sh"
	case "ruby":
		return ".rb"
	case "php":
		return ".php"
	}
	return ""
}

func trimVersion(name string) string {
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= '0' && c <= '9') || c == '.' {
			return name[:i]
		}
	}
	return name
}

func writeAtomic(files FileSystem, path string, data []byte, perm fs.FileMode) error {
	f, err := files.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".nocomment-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		files.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		files.Remove(tmp)
		return err
	}
	if err := files.Chmod(tmp, perm); err != nil {
		files.Remove(tmp)
		return err
	}
	if err := files.Rename(tmp, path); err != nil {
		files.Remove(tmp)
		return err
	}
	return nil
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
