package storageclient

import (
	"fmt"
	"log"

	"Raft-TFG/file-creator_v1.0/settings"
)

const (
	minioType string = "minio"
)

func GetStorageClient(settings *settings.Settings) (IStorageClient, error) {
	log.Default().Printf("Connecting to %v object storage...", settings.StorageType)

	switch settings.StorageType {
	case minioType:
		return newMinioclient(settings), nil
	default:
		return nil, fmt.Errorf("The storage type is not supported\n")
	}
}
