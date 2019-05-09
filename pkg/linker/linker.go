// Package linker runs very specific logic to simply link together javascript
// files into bundles. It traverses specific entrypoints to find packages to
// import.
//
// Architecture:
// - Traverse entrypoints and recursively find all the files given import syntax.
// - Load these files together as bundles.
package linker

import (
	"log"
	"sort"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"

	"github.com/coldog/jsbld/pkg/compiler"
	"github.com/coldog/jsbld/pkg/util"
)

type File struct {
	compiler.Object
	Entrypoints []string
}

type Files map[string]File

func (f Files) Add(file, entrypoint string) {
	f[file] = File{Entrypoints: append(f[file].Entrypoints, entrypoint)}
}

func (f Files) SetObject(file string, o compiler.Object) {
	f[file] = File{Entrypoints: f[file].Entrypoints, Object: o}
}

func (f Files) Keys() []string {
	keys := []string{}
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type Chunk struct {
	Files      Files
	Entrypoint string
	Loads      []string
}


func (c Chunk) Output() string {
	h := sha256.New()
	for _, f := range c.Files {
		h.Write([]byte(f.Hash))
	}
	hash := hex.EncodeToString(h.Sum(nil))
	prefix := strings.Split(filepath.Base(c.Entrypoint), ".")[0]
	return prefix + "-" + hash + ".js"
}

type Bundle struct {
	Root        string
	Files       Files
	Entrypoints []string
	Chunks      []*Chunk
}

func (b *Bundle) Write() error {
	popd := util.Pushd(b.Root)
	defer popd()

	for _, chunk := range b.Chunks {
		log.Printf("writing: %s", chunk.Output())
		var err error
		if chunk.Entrypoint != "" {
			err = bundle(chunk.Files, chunk.Entrypoint, chunk.Output(), chunk.Loads)
		} else {
			err = bundleChunk(chunk.Files, chunk.Output())
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bundle) Find() error {
	popd := util.Pushd(b.Root)
	defer popd()

	b.Files = Files{}
	for _, entrypoint := range b.Entrypoints {
		b.Files.Add(entrypoint, entrypoint)

		if err := parse(b.Files, entrypoint, entrypoint); err != nil {
			return err
		}
	}
	return nil
}

// Loads files into the files map and traverses child dependencies.
func parse(files Files, file, entrypoint string) error {
	o, err := compiler.ReadObjectFile(file)
	if err != nil {
		return err
	}
	files.SetObject(file, o)

	for _, require := range o.Imports {
		if _, ok := files[require]; ok {
			continue
		}

		log.Printf("resolve: %s", require)
		if err != nil {
			return err
		}

		files.Add(require, entrypoint)
		err = parse(files, require, entrypoint)
		if err != nil {
			return err
		}
	}
	return nil
}
