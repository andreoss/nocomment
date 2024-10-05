package lang

import (
	"path/filepath"
	"strings"
	"sync"

	"nocomment/internal/lexer"
)

type Language interface {
	Name() string
	Extensions() []string
	Tokenize(src []byte) ([]lexer.Token, error)
	IsComment(name string) bool
}

type Registry struct {
	mu     sync.RWMutex
	byExt  map[string]Language
	byName map[string]Language
}

func NewRegistry() *Registry {
	return &Registry{byExt: map[string]Language{}, byName: map[string]Language{}}
}

func (r *Registry) Register(l Language) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byName[l.Name()] = l
	for _, ext := range l.Extensions() {
		r.byExt[NormalizeExt(ext)] = l
	}
}

func (r *Registry) ByName(name string) (Language, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.byName[name]
	return l, ok
}

func (r *Registry) ByExtension(path string) (Language, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.byExt[NormalizeExt(filepath.Ext(path))]
	return l, ok
}

func (r *Registry) Languages() []Language {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := map[string]bool{}
	out := make([]Language, 0, len(r.byName))
	for _, l := range r.byName {
		if seen[l.Name()] {
			continue
		}
		seen[l.Name()] = true
		out = append(out, l)
	}
	return out
}

func NormalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}
