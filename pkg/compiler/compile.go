package compiler

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coldog/jsbld/pkg/resolve"
	"github.com/coldog/jsbld/pkg/util"
)

var (
	BabelCompiler   = "babel $1 --compact=true --config-file=./.babelrc --out-file=$2"
	DefaultCompiler = "cp $1 $2"
)

var Compilers = map[string]string{
	"js":  BabelCompiler,
	"jsx": BabelCompiler,
	"tsx": BabelCompiler,
	"ts":  BabelCompiler,
	"*":   DefaultCompiler,
}

func getCompiler(name, srcFile, dstFile string) []string {
	spl := strings.Split(name, ".")
	ext := spl[len(spl)-1]
	c := Compilers[ext]
	if c == "" {
		c = Compilers["*"]
	}
	c = strings.Replace(c, "$1", srcFile, 1)
	c = strings.Replace(c, "$2", dstFile, 1)
	return strings.Fields(c)
}

func isJS(name string) bool {
	for _, ext := range resolve.Extensions {
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
	os.MkdirAll(filepath.Dir(dstFile), 0777)

	object := Object{Filename: dstFile}
	{
		prev, _ := ReadObjectFile(dstFile)
		h, err := hash(srcFile)
		if err != nil {
			return err
		}
		if prev.Hash == h {
			return nil
		}
		object.Hash = h
	}

	compiler := getCompiler(file, srcFile, dstFile)
	cmd := exec.Command(compiler[0], compiler[1:]...)

	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	if isJS(file) {
		imps, err := compileImports(srcFile, dstFile)
		if err != nil {
			return err
		}
		object.Imports = imps
	}

	return WriteObjectFile(object)
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
	return errs.first()
}
