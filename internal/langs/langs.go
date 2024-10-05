package langs

import (
	antlr4 "github.com/antlr4-go/antlr/v4"

	"nocomment/internal/lang"
	"nocomment/internal/lang/golang"
	"nocomment/internal/lexer"
	"nocomment/internal/lexer/generated/cpp"
	"nocomment/internal/lexer/generated/html"
	"nocomment/internal/lexer/generated/java"
	"nocomment/internal/lexer/generated/kotlin"
	"nocomment/internal/lexer/generated/postgresql"
	"nocomment/internal/lexer/generated/sqlite"
	"nocomment/internal/lexer/generated/toml"
)

func Default() *lang.Registry {
	r := lang.NewRegistry()
	r.Register(golang.New())
	r.Register(lang.NewGeneric("java", []string{".java"}, javaTokens,
		[]string{"COMMENT", "LINE_COMMENT"}))
	r.Register(lang.NewGeneric("c", []string{".c", ".h"}, cppTokens,
		[]string{"BlockComment", "LineComment"}))
	r.Register(lang.NewGeneric("sql", []string{".sql"}, sqliteTokens,
		[]string{"MULTILINE_COMMENT", "SINGLE_LINE_COMMENT"}))
	r.Register(lang.NewGeneric("html", []string{".html", ".htm"}, htmlTokens,
		[]string{"HTML_COMMENT", "HTML_CONDITIONAL_COMMENT"}))
	r.Register(lang.NewGeneric("xml", []string{".xml"}, htmlTokens,
		[]string{"HTML_COMMENT", "HTML_CONDITIONAL_COMMENT"}))
	r.Register(lang.NewGeneric("kotlin", []string{".kt", ".kts"}, kotlinTokens,
		[]string{"DelimitedComment", "Inside_Comment", "LineComment", "StrExpr_Comment"}))
	r.Register(lang.NewGeneric("sqlite", []string{".sqlite", ".sqlite3"}, sqliteTokens,
		[]string{"MULTILINE_COMMENT", "SINGLE_LINE_COMMENT"}))
	r.Register(lang.NewGeneric("toml", []string{".toml"}, tomlTokens,
		[]string{"COMMENT"}))
	r.Register(lang.NewGeneric("cpp", []string{".cpp", ".cc", ".cxx", ".hpp", ".hh"}, cppTokens,
		[]string{"BlockComment", "LineComment"}))
	r.Register(lang.NewGeneric("postgresql", []string{".pgsql"}, postgresqlTokens,
		[]string{"BlockComment", "LineComment", "UnterminatedBlockComment"}))
	registerScanners(r)
	return r
}

func javaTokens(src []byte) ([]lexer.Token, error) {
	l := java.NewJavaLexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}

func htmlTokens(src []byte) ([]lexer.Token, error) {
	l := html.NewHTMLLexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}

func kotlinTokens(src []byte) ([]lexer.Token, error) {
	l := kotlin.NewKotlinLexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}

func sqliteTokens(src []byte) ([]lexer.Token, error) {
	l := sqlite.NewSQLiteLexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}

func tomlTokens(src []byte) ([]lexer.Token, error) {
	l := toml.NewTomlLexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}

func cppTokens(src []byte) ([]lexer.Token, error) {
	l := cpp.NewCPP14Lexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}

func postgresqlTokens(src []byte) ([]lexer.Token, error) {
	l := postgresql.NewPostgreSQLLexer(antlr4.NewInputStream(string(src)))
	return lexer.Collect(src, l.NextToken, l.SymbolicNames), nil
}
