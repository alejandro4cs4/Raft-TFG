package main

import (
	"fmt"
	"os"
	"strings"
)

const directoryPath string = "/home/alejandroc/etcd"

func main() {
	listDirectory(directoryPath, 0)
}

func listDirectory(path string, listLevel int) {
	directoryEntries, err := os.ReadDir(path)
	checkError(err)

	for _, directoryEntry := range directoryEntries {
		for i := 0; i < listLevel; i++ {
			fmt.Print("  ")
		}
		fmt.Printf("%s\n", directoryEntry.Name())

		if directoryEntry.IsDir() {
			listDirectory(strings.Join([]string{path, directoryEntry.Name()}, "/"), listLevel+1)
		}
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
