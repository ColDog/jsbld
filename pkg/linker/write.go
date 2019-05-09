package linker

import (
	"bufio"
	"io"
	"os"
	"encoding/json"
)

const header = "\"use strict\";\n(function() {\n"
const footer = "})();\n"

func bundle(files Files, entry, output string, loads []string) error {
	f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	_, err = w.WriteString(header)
	if err != nil {
		return err
	}
	_, err = w.WriteString(runtime)
	if err != nil {
		return err
	}
	err = writeFiles(files, w)
	if err != nil {
		return err
	}
	err = writeStart(w, entry, loads)
	if err != nil {
		return err
	}
	_, err = w.WriteString(footer)
	if err != nil {
		return err
	}
	return w.Flush()
}

func bundleChunk(files Files, output string) error {
	f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	_, err = w.WriteString(header)
	if err != nil {
		return err
	}
	err = writeFiles(files, w)
	if err != nil {
		return err
	}
	_, err = w.WriteString(footer)
	if err != nil {
		return err
	}
	return w.Flush()
}

func writeStart( w *bufio.Writer, entrypoint string, chunkPaths []string) error {
	data, err := json.Marshal(chunkPaths)
	if err != nil {
		return err
	}
	_, err = w.WriteString("start("+ string(data) +", \"" + entrypoint + "\")")
	return err
}

func writeFiles(files Files, w *bufio.Writer) error {
	for _, file := range files.Keys() {
		fm, err := os.Open(file)
		if err != nil {
			return err
		}
		w.WriteString("window.__modules__[\"" + file + "\"] = function(module, exports, require) {\n")
		_, err = io.Copy(w, fm)
		if err != nil {
			return err
		}
		w.WriteString("};\n")
	}
	return nil
}
