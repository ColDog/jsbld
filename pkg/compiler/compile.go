package compiler

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coldog/bld/pkg/util"
)

var extensions = []string{"js", "jsx", "tsx", "ts"}

func canCompile(name string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}



// compileFile is very simple in that it takes a file and writes a compiled
// file.
func compileFile(src, dst, file string) error {
	srcFile := filepath.Join(src, file)
	dstFile := filepath.Join(dst, src, file)
	os.MkdirAll(filepath.Dir(dstFile), 0700)

	var cmd *exec.Cmd
	if canCompile(file) {
		cmd = exec.Command(
			"babel", srcFile,
			"--compact=true",
			"--config-file", "./.babelrc",
			"--out-file", dstFile,
		)
	} else {
		cmd = exec.Command("cp", srcFile, dstFile)
	}

	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	if canCompile(file) {
		imps, err := compileImports(dstFile)
		if err != nil {
			return err
		}

		data, err := json.Marshal(imps)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(dstFile+".o", data, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func hash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	hash := hex.EncodeToString(h.Sum(nil))
	return hash, nil
}

func loadState() map[string]string {
	f, err := os.Open("bld.json")
	if err != nil {
		return map[string]string{}
	}
	defer f.Close()

	m := map[string]string{}
	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		log.Printf("failed to read state: %v", err)
		return map[string]string{}
	}
	return m
}

func saveState(state map[string]string) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	ioutil.WriteFile("bld.json", data, 0700)
}

type errList struct {
	lock sync.Mutex
	errs []error
}

func (e *errList) push(err error) {
	e.lock.Lock()
	e.errs = append(e.errs, err)
	e.lock.Unlock()
}

func (e *errList) first() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if len(e.errs) == 0 {
		return nil
	}
	return e.errs[0]
}

func Compile(root, dst string, srcs []string) error {
	popd := util.Pushd(root)
	defer popd()

	state := loadState()

	concurrency := 10
	os.MkdirAll(dst, 0700)

	wg := sync.WaitGroup{}
	errs := &errList{}
	paths := make(chan struct {
		path string
		src  string
	}, concurrency)

	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(i int) {
			defer wg.Done()

			for path := range paths {
				t1 := time.Now()
				err := compileFile(path.src, dst, path.path)
				if err != nil {
					errs.push(err)
				}
				log.Printf("compile(%d): %s -- %v (%v)", i, path, err, time.Since(t1))
			}
		}(i)
	}

	for _, src := range srcs {
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			h, err := hash(path)
			if err != nil {
				return err
			}
			if state[path] == h {
				log.Printf("compile: %s -- (cached)", path)
				return nil
			}
			state[path] = h

			rel, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}

			paths <- struct {
				path string
				src  string
			}{
				path: rel,
				src:  src,
			}
			return nil
		})
		if err != nil {
			errs.push(err)
		}
	}

	close(paths)
	wg.Wait()

	if errs.first() == nil {
		saveState(state)
	}
	return errs.first()
}
