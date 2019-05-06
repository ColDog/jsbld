// Package linker runs very specific logic to simply link together javascript
// files into bundles. It traverses specific entrypoints to find packages to
// import.
//
// Architecture:
// - Traverse entrypoints and recursively find all the files given import syntax.
// - Load these files together as bundles.
package linker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"bufio"
	"os"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/coldog/bld/pkg/util"
)

var extensions = []string{"js", "jsx", "tsx", "ts"}



const runtime = `
var cache = {};
var modules = {};
function require(name) {
  if (cache[name]) {
    return cache[name].exports;
  }
  var module = {
    name: name,
    exports: {}
  };
  modules[name](module, module.exports, require);
  cache[name] = module;
  return module.exports;
}
`

func bundle(files map[string]string, entry, output string) error {
	f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	w.WriteString("\"use strict\";\n")
	w.WriteString("(function() {\n")
	w.WriteString(runtime)

	for name, file := range files {
		fm, err := os.Open(file)
		if err != nil {
			return err
		}
		w.WriteString("\n\n/* " + file + " */\n")
		w.WriteString("modules[\"" + name + "\"] = function(module, exports, require) {")
		_, err = io.Copy(w, fm)
		if err != nil {
			return err
		}
		w.WriteString("};\n\n")
	}

	w.WriteString("require(\"" + entry + "\");\n")
	w.WriteString("})();\n")
	return w.Flush()
}

func Link(root, entrypoint, output string) error {
	popd := util.Pushd(root)
	defer popd()

	resolved, err := resolve(root, entrypoint)
	if err != nil {
		return err
	}

	// Map of files to resolved location on disk.
	files := map[string]string{
		entrypoint: resolved,
	}

	if err := parse(files, resolved); err != nil {
		return err
	}
	return bundle(files, entrypoint, output)
}


// Loads files into the files map and traverses child dependencies.
func parse(files map[string]string, file string) error {
	data, err := ioutil.ReadFile(file + ".o")
	if err != nil {
		return err
	}
	requires := []string{}
	err = json.Unmarshal(data, &requires)
	if err != nil {
		return err
	}

	for _, require := range requires {
		if _, ok := files[require]; ok {
			continue
		}

		resolved, err := resolve(filepath.Dir(file), require)
		log.Printf("resolve: %s", resolved)

		if err != nil {
			return err
		}

		files[require] = resolved
		err = parse(files, resolved)
		if err != nil {
			return err
		}
	}
	return nil
}
