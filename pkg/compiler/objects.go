package compiler

import (
	"encoding/hex"
	"io"
	"os"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
)

type Object struct {
	Filename string
	Hash     string
	Imports  []string
}

func WriteObjectFile(o Object) error {
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(o.Filename + ".o", data, 0777)
}

func ReadObjectFile(file string) (Object, error) {
	data, err := ioutil.ReadFile(file + ".o")
	if err != nil {
		return Object{}, err
	}
	o := Object{}
	err = json.Unmarshal(data, &o)
	return o, err
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

