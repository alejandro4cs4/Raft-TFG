package pfslib

import (
	"log"
	"strings"
)

// PfsOpen helper functions //

func printOpenNotFound(pathComponents []string) {
	pathnameNotFound := strings.Join(pathComponents, "/")

	log.Default().Printf("[pfslib]: PfsOpen(): no such file or directory \"%s\"\n", pathnameNotFound)

}

func printOpenNotRegular(absolutePath string) {
	log.Default().Printf("[pfslib]: PfsOpen(): \"%s\" is not a regular file\n", absolutePath)
}

func printOpenOutOfMemory() {
	log.Default().Println("[pfslib]: PfsOpen(): out of memory")
}

// PfsRead helper functions //

func printReadInvalidDescriptor(invalidDescriptor int) {
	log.Default().Printf("[pfslib]: PfsRead(): the file descriptor \"%d\" is invalid\n", invalidDescriptor)
}

func printReadNotOpenedDescriptor(invalidDescriptor int) {
	log.Default().Printf("[pfslib]: PfsRead(): no opened file associated to the specified descriptor \"%d\"\n", invalidDescriptor)
}
