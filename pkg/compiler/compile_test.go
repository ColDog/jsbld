package compiler

import (
	"testing"
)

func TestExample(t *testing.T) {
	err := Compile("../../example", "dst", []string{"src", "node_modules"})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}

// func TestCompileImports(t *testing.T) {
// 	imps, err := compileImports("../../example/node_modules/react/index.js")
// 	if err != nil {
// 		t.Fatalf("failed: %v", err)
// 	}
// 	if fmt.Sprintf("%+v", imps) != "[./cjs/react.production.min.js]" {
// 		t.Fatalf("wrong imports: '%+v'", imps)
// 	}
// }
