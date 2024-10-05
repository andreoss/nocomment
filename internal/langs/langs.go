package langs

import (
	"nocomment/internal/lang"
	"nocomment/internal/lang/golang"
)

func Default() *lang.Registry {
	r := lang.NewRegistry()
	r.Register(golang.New())
	return r
}
