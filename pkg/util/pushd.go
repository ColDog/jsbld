package util

import "os"

func Pushd(root string) func() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(root)
	if err != nil {
		panic(err)
	}

	return func() {
		err := os.Chdir(wd)
		if err != nil {
			panic(err)
		}
	}
}
