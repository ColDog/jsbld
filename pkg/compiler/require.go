package compiler

import (
	"os"
	"io"
	"bytes"
	"path/filepath"
	"bufio"
	"log"

	"github.com/coldog/jsbld/pkg/resolve"
)

// An efficient parsing function that does the following steps to setup require
// statements for linking. It maintains a bit of ugly code below for a fast
// implementation.
//
// 1. Parse all require(...) calls.
// 2. Rewrite require statements with the full path:
//		require('react') -> require('node_modules/react').
// 3. Returns full paths of all required files.
func compileImports(srcFile, dstFile string) ([]string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	f, err := os.OpenFile(dstFile, os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var imports []string
	var prev rune
	var inR bool

	rd := bufio.NewReader(f)
	for {
		c, _, err := rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return imports, err
		}
		switch c {
		case 'r':
			if !inR {
				inR = true
			} else if inR && prev != 'i' {
				inR = false
			}
			buf.WriteRune(c)
		case 'e':
			if prev != 'r' {
				inR = false
			}
			buf.WriteRune(c)
		case 'q':
			if prev != 'e' {
				inR = false
			}
			buf.WriteRune(c)
		case 'u':
			if prev != 'q' {
				inR = false
			}
			buf.WriteRune(c)
		case 'i':
			if prev != 'u' {
				inR = false
			}
			buf.WriteRune(c)
		case '(':
			if prev != 'e' {
				inR = false
			}
			buf.WriteRune(c)
		case '\'', '"':
			if prev != '(' {
				inR = false
			}
			buf.WriteRune(c)
			if inR {
				imp, err := rd.ReadString(byte(c))
				if err != nil {
					if err == io.EOF {
						break
					}
					return imports, err
				}
				imp = imp[:len(imp)-1]

				fullPath, err := resolve.Resolve(filepath.Dir(srcFile), imp)
				if err == nil {
					imports = append(imports, fullPath)
					buf.WriteString(fullPath)
				} else {
					log.Printf("failed to resolve: %s in %s -- %v", imp, filepath.Dir(srcFile), err)
					// If we couldn't resolve the path correctly we just leave
					// this and don't log it as a dependent import. This means
					// that if this require() is called that in the browser it
					// will fail.
					buf.WriteString(imp)
				}
				buf.WriteRune(c)
				inR = false
			}
		default:
			buf.WriteRune(c)
		}
		prev = c
	}

	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	if _, err := buf.WriteTo(f); err != nil {
		return nil, err
	}
	return imports, nil
}
