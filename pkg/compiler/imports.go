package compiler

import (
	"os"
	"fmt"
	"bufio"
	"io"
	"bytes"
	"strings"
	"path/filepath"
	"encoding/json"
)

// Resolve implements a basic node resolution algorithm. It returns a relative
// path to a file.
func resolve(root, name string) (string, error) {
	if !(strings.HasPrefix(name, "../") || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "/")) {
		name = filepath.Join("node_modules", name)
	} else {
		name = filepath.Join(root, name)
	}

	st, err := os.Stat(name)
	if err != nil {
		for _, ext := range extensions {
			name = name + "." + ext
			st, err = os.Stat(name)
			if err == nil {
				break
			}
		}
	}

	if st == nil {
		return "", fmt.Errorf("could not resolve: \"%s\"", name)
	}

	if st.IsDir() {
		_, err := os.Stat(filepath.Join(name, "package.json"))
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}

		main := "index.js"
		if err == nil {
			f, err := os.Open(filepath.Join(name, "package.json"))
			if err != nil {
				return "", err
			}

			m := struct {
				Main string `json:"main"`
			}{}
			json.NewDecoder(f).Decode(&m)
			f.Close()

			if m.Main != "" {
				main = m.Main
			}
		}

		return filepath.Join(name, main), nil
	}

	return name, nil
}

func compileImports(file string) ([]string, error) {
	buf := bytes.NewBuffer(make([]byte, 1024))
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var imports []string
	var prev byte
	var inR bool

	rd := bufio.NewReader(f)
	for {
		c, err := rd.ReadByte()
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
			buf.WriteByte(c)
		case 'e':
			if prev != 'r' {
				inR = false
			}
			buf.WriteByte(c)
		case 'q':
			if prev != 'e' {
				inR = false
			}
			buf.WriteByte(c)
		case 'u':
			if prev != 'q' {
				inR = false
			}
			buf.WriteByte(c)
		case 'i':
			if prev != 'u' {
				inR = false
			}
			buf.WriteByte(c)
		case '(':
			if prev != 'e' {
				inR = false
			}
			buf.WriteByte(c)
		case '\'', '"':
			if prev != '(' {
				inR = false
			}
			buf.WriteByte(c)
			if inR {
				imp, err := rd.ReadString(c)
				if err != nil {
					if err == io.EOF {
						break
					}
					return imports, err
				}
				imp = imp[:len(imp)-1]

				imports = append(imports, imp)
				inR = false
			}
			buf.WriteByte(c)
		}
		prev = c
	}

	return imports, nil
}
