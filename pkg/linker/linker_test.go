package linker

import (
	"testing"

	"github.com/coldog/jsbld/pkg/compiler"
)

func TestExample(t *testing.T) {
	compiler.Compile("../../example", "dst", []string{"src", "node_modules"})

	b := &Bundle{
		Root:        "../../example/dst",
		Entrypoints: []string{"./src/index.js"},
	}
	err := b.Find()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	err = StandardBundler(b)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	err = b.Write()
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}
