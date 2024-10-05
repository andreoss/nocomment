package langs

import (
	"nocomment/internal/lang"
	"nocomment/internal/lexer"
	"nocomment/internal/scan"
)

func scannerTokens(s scan.Syntax) func([]byte) ([]lexer.Token, error) {
	return func(src []byte) ([]lexer.Token, error) {
		ranges := scan.Ranges(src, s)
		tokens := make([]lexer.Token, 0, len(ranges))
		for _, r := range ranges {
			if r.Stop <= r.Start {
				continue
			}
			tokens = append(tokens, lexer.Token{Name: "COMMENT", Start: r.Start, Stop: r.Stop - 1})
		}
		return tokens, nil
	}
}

func registerScanners(r *lang.Registry) {
	r.Register(lang.NewGeneric("python", []string{".py"}, scannerTokens(scan.Syntax{
		Line: []string{"#"},
		Strings: []scan.StringRule{
			{Open: `"""`, Close: `"""`, Multiline: true},
			{Open: `'''`, Close: `'''`, Multiline: true},
			{Open: `"`, Close: `"`, Escape: true},
			{Open: `'`, Close: `'`, Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("csharp", []string{".cs"}, scannerTokens(scan.Syntax{
		Line:  []string{"//"},
		Block: [][2]string{{"/*", "*/"}},
		Strings: []scan.StringRule{
			{Open: `$@"`, Close: `"`, Doubled: true, Multiline: true},
			{Open: `@"`, Close: `"`, Doubled: true, Multiline: true},
			{Open: `$"`, Close: `"`, Escape: true},
			{Open: `"""`, Close: `"""`, Multiline: true},
			{Open: `"`, Close: `"`, Escape: true},
			{Open: `'`, Close: `'`, Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("php", []string{".php"}, scannerTokens(scan.Syntax{
		Line:  []string{"//", "#"},
		Block: [][2]string{{"/*", "*/"}},
		Strings: []scan.StringRule{
			{Open: `"`, Close: `"`, Escape: true},
			{Open: `'`, Close: `'`, Escape: true},
			{Open: "`", Close: "`", Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("swift", []string{".swift"}, scannerTokens(scan.Syntax{
		Line:  []string{"//"},
		Block: [][2]string{{"/*", "*/"}},
		Nest:  true,
		Strings: []scan.StringRule{
			{Open: `"""`, Close: `"""`, Multiline: true},
			{Open: `"`, Close: `"`, Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("ruby", []string{".rb"}, scannerTokens(scan.Syntax{
		Line:           []string{"#"},
		BlockLineStart: [][2]string{{"=begin", "=end"}},
		Strings: []scan.StringRule{
			{Open: `"`, Close: `"`, Escape: true},
			{Open: `'`, Close: `'`, Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("shell", []string{".sh", ".bash"}, scannerTokens(scan.Syntax{
		Line:         []string{"#"},
		LineBoundary: true,
		Strings: []scan.StringRule{
			{Open: `"`, Close: `"`, Escape: true},
			{Open: `'`, Close: `'`},
			{Open: "`", Close: "`", Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("jsonc", []string{".jsonc", ".json"}, scannerTokens(scan.Syntax{
		Line:  []string{"//"},
		Block: [][2]string{{"/*", "*/"}},
		Strings: []scan.StringRule{
			{Open: `"`, Close: `"`, Escape: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("yaml", []string{".yaml", ".yml"}, scannerTokens(scan.Syntax{
		Line:         []string{"#"},
		LineBoundary: true,
		Strings: []scan.StringRule{
			{Open: `"`, Close: `"`, Escape: true},
			{Open: `'`, Close: `'`, Doubled: true},
		},
	}), []string{"COMMENT"}))

	r.Register(lang.NewGeneric("markdown", []string{".md", ".markdown"}, scannerTokens(scan.Syntax{
		Block: [][2]string{{"<!--", "-->"}},
	}), []string{"COMMENT"}))
}
