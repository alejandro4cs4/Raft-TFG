package pfslib

import (
	"log"

	"raft-tfg.com/alejandroc/pfslib/globals"
	"raft-tfg.com/alejandroc/pfslib/metaclient"
	"raft-tfg.com/alejandroc/pfslib/storeclient"
	"raft-tfg.com/alejandroc/pfslib/utils"
)

func PfsInit() {
	utils.ReadSettings()

	// Init etcd client
	log.Default().Printf("[pfslib]: Connecting to etcd metadata storage...")
	metaclient.GetEtcdClient()
	log.Default().Printf("[pfslib]: Connected to etcd metadata storage successfully")

	// Init MinIO client
	log.Default().Printf("[pfslib]: Connecting to MinIO object storage...")
	storeclient.GetMinioClient()
	log.Default().Printf("[pfslib]: Connected to MinIO object storage successfully")

	// Init openfds table
	initOpenfds()
}

func initOpenfds() {
	globals.Openfds = []globals.Openfd{}
}

// Tries to open the pathname file and return its associated file descriptor
func PfsOpen(pathname string) (int, error) {
	// Solve file pathname contacting etcd
	// - start from root (/) if absolute pathname
	// - start from current directory if relative pathname

	// If file is not regular file (F) -> return error

	// Get file's UUID from its etcd value

	// Look for a free entry in openfds table and store the file UUID

	// Return openfds index for the recently opened file

	return -1, nil
}
