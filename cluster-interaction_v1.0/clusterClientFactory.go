package main

import "fmt"

func getClusterClient(clusterType string) (IClusterClient, error) {
	if clusterType == "etcd" {
		return newEtcdClient(), nil
	}

	if clusterType == "tikv" {
		return newTikvClient(), nil
	}

	return nil, fmt.Errorf("The specified cluster type is not supported\n")
}