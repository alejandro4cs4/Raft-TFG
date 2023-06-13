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

// PfsLseek helper functions //
func printInvalidOffset(invalidOffset int64) {
	log.Default().Printf("[pfslib]: Lseek(): offset \"%d\" is invalid, must be greater than 0\n", invalidOffset)
}

func printInvalidWhence(invalidWhence int) {
	log.Default().Printf("[pfslib]: Lseek(): whence \"%d\" is invalid, check go os package to see available values\n", invalidWhence)
}

// Other helper functions //
func printClosedPfsFile(operationName string) {
	log.Default().Printf("[pfslib]: %s(): the PfsFile is already closed\n", operationName)
}

func printInvalidBuffer(operationName string) {
	log.Default().Printf("[pfslib]: %s(): the provided buffer must be non nil\n", operationName)
}
