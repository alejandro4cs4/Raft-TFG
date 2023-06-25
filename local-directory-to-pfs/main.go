package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"raft-tfg.com/alejandroc/pfslib"
)

const directoryToCopy string = "/home/alejandroc/etcd"

func main() {
	pfslib.PfsInit()

	defer pfslib.PfsEnd()

	startTime := time.Now()

	pfslib.PfsMkdirAll(directoryToCopy, 666)

	copyDirectory(directoryToCopy)

	elapsedTime := time.Since(startTime)

	fmt.Printf("It took %d milliseconds / %.2f seconds / %.2f minutes to copy \"%s\" to the pfs\n", elapsedTime.Milliseconds(), elapsedTime.Seconds(), elapsedTime.Minutes(), directoryToCopy)
}

func copyContainerDirectories(path string) {
	pathComponents := strings.Split(path, "/")
	pathComponents = pathComponents[:len(pathComponents)-1]

	for index := range pathComponents {
		directoryToCreatePath := strings.Join(pathComponents[:index+1], "/")

		pfslib.PfsMkdir(directoryToCreatePath, 666)
	}
}

func copyDirectory(path string) {
	if path == "/home/alejandroc/etcd-cluster" || path == "/home/alejandroc/minio" {
		fmt.Println("Hago return")
		return
	}

	pfslib.PfsMkdir(path, 666)

	directoryEntries, err := os.ReadDir(path)
	checkError(err)

	for _, directoryEntry := range directoryEntries {
		currentPath := strings.Join([]string{path, directoryEntry.Name()}, "/")

		if directoryEntry.IsDir() {
			copyDirectory(currentPath)
		} else if directoryEntry.Type().IsRegular() {
			pfslib.PfsCreate(currentPath)
		}
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
