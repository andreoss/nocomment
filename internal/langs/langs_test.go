package langs

import "testing"

func TestDefault(t *testing.T) {
	r := Default()
	if _, ok := r.ByName("go"); !ok {
		t.Fatal("go not registered")
	}
	if _, ok := r.ByExtension("main.go"); !ok {
		t.Fatal(".go not registered")
	}
}
