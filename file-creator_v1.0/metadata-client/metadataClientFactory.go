package metadataclient

import (
	"fmt"
	"log"

	"Raft-TFG/file-creator_v1.0/settings"
)

const (
	etcdType string = "etcd"
	tikvType        = "tikv"
)

func GetMetadataClient(settings *settings.Settings) (IMetadataClient, error) {
	log.Default().Printf("Connecting to %v metadata storage...", settings.MetadataType)

	switch settings.MetadataType {
	case etcdType:
		return newEtcdClient(settings), nil
	case tikvType:
		return newTikvClient(settings), nil
	default:
		return nil, fmt.Errorf("The metadata type is not supported\n")
	}
}
