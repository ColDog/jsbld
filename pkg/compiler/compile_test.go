package compiler

import (
	"testing"
	"fmt"
)

func TestExample(t *testing.T) {
	err := Compile("../../example", "dst", []string{"src", "node_modules"})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}

func TestCompileImports(t *testing.T) {
	imps, err := compileImports("testdata/imports.js")
	if fmt.Sprintf("%+v", imps) != "[react react-dom ./cjs/react.production.min.js]" {
		t.Fatalf("wrong imports: '%+v'", imps)
	}
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}
