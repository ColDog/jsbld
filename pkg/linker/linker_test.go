package linker

import (
	"testing"
	"fmt"
	"github.com/coldog/bld/pkg/compiler"
)

func TestExample(t *testing.T) {
	compiler.Compile("../../example", "dst", []string{"src", "node_modules"})
	err := Link("../../example/dst", "./src/index.js", "./bundle.js")
	fmt.Printf("err: %v", err)
}
