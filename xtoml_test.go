package xtoml

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
)

const (
	tst1 = `
[[data]]
val1 = 0

#[[data]]
#val1 = 1

#[[data]]
#val1 = 2
`
)

func TestTree(t *testing.T) {
	rdr := strings.NewReader(tst1)
	tt, err := toml.LoadReader(rdr)
	if err != nil {
		t.Fatal(err)
	}
	v := tt.Get("data")
	if v == nil {
		t.Log("path not found")
		return
	}
	t.Logf("%T\n", v)
	tt1, ok := v.([]*toml.Tree)
	if !ok {
		t.Fatal("wrong type")
	}
	for _, ttt := range tt1 {
		t.Log(ttt.Keys())
	}
}
