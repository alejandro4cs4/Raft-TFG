package main

import (
	"fmt"
	"path/filepath"

	"raft-tfg.com/alejandroc/pfslib"
)

func main() {
	pfslib.PfsInit()

	fmt.Println()

	evaluatedPath, _ := filepath.EvalSymlinks("./../pfsclient/../pfsclient/settings.yaml")
	canonicalPath, _ := filepath.Abs(evaluatedPath)

	fmt.Printf("%v\n", canonicalPath)
}
