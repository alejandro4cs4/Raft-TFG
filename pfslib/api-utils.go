package pfslib

import (
	"context"
	"log"
	"strings"

	"github.com/minio/minio-go/v7"

	"raft-tfg.com/alejandroc/pfslib/globals"
	"raft-tfg.com/alejandroc/pfslib/utils"
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

func printReadNotOpenedDescriptor(invalidDescriptor int) {
	log.Default().Printf("[pfslib]: PfsRead(): no opened file associated to the specified descriptor \"%d\"\n", invalidDescriptor)
}

// PfsLseek helper functions //
func getFileSize(fileDescriptor int) int64 {
	minioObjectInfo, err := storeClient.StatObject(context.Background(), globals.MinioBucket, globals.Openfds[fileDescriptor].Uuid, minio.StatObjectOptions{})
	utils.CheckError(err)

	return minioObjectInfo.Size
}

func printInvalidOffset(invalidOffset int64) {
	log.Default().Printf("[pfslib]: PfsLseek(): offset \"%d\" is invalid, must be greater than 0\n", invalidOffset)
}

func printInvalidWhence(invalidWhence int) {
	log.Default().Printf("[pfslib]: PfsLseek(): whence \"%d\" is invalid, check go os package to see available values\n", invalidWhence)
}

// Other helper functions //

func printInvalidDescriptor(invalidDescriptor int, operationName string) {
	log.Default().Printf("[pfslib]: Pfs%s(): the file descriptor \"%d\" is invalid\n", operationName, invalidDescriptor)
}

func printInvalidBuffer(operationName string) {
	log.Default().Printf("[pfslib]: Pfs%s(): the provided buffer must be non nil\n", operationName)
}
