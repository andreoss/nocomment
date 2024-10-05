package process

import (
	"errors"

	"nocomment/internal/filter"
	"nocomment/internal/lang"
)

type Processor struct {
	Registry *lang.Registry
	Language string
}

func (p Processor) Process(path string, src []byte) ([]byte, error) {
	l, err := p.lookup(path)
	if err != nil {
		return nil, err
	}
	return filter.Strip(l, src)
}

func (p Processor) lookup(path string) (lang.Language, error) {
	if p.Language != "" {
		if l, ok := p.Registry.ByName(p.Language); ok {
			return l, nil
		}
		return nil, errors.New("unknown language: " + p.Language)
	}
	if path == "" {
		return nil, errors.New("standard input requires -lang")
	}
	if l, ok := p.Registry.ByExtension(path); ok {
		return l, nil
	}
	return nil, errors.New("unsupported file: " + path)
}
