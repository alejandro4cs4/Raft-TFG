package main

import "fmt"

func getClusterClient(settings *Settings) (IClusterClient, error) {
	if settings.ClusterType == "etcd" {
		return newEtcdClient(settings), nil
	}

	if settings.ClusterType == "tikv" {
		return newTikvClient(settings), nil
	}

	return nil, fmt.Errorf("The specified cluster type is not supported\n")
}