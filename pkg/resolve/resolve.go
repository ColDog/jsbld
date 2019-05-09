package resolve

import (
	"strings"
	"path/filepath"
	"os"
	"fmt"
	"encoding/json"
)

var Extensions = []string{"js", "jsx", "tsx", "ts"}

// Resolve implements a basic node resolution algorithm. It returns a relative
// path to a file.
func Resolve(root, name string) (string, error) {
	if !(strings.HasPrefix(name, "../") || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "/")) {
		name = filepath.Join("node_modules", name)
	} else {
		name = filepath.Join(root, name)
	}

	st, err := os.Stat(name)
	if err != nil {
		for _, ext := range Extensions {
			st, err = os.Stat(name + "." + ext)
			if err == nil {
				name = name + "." + ext
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
