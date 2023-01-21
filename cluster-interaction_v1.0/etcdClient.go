package main

import (
	"context"
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	client *clientv3.Client
}

func newEtcdClient() IClusterClient {
	config := clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	}

	errChan := make(chan error, 1)
	cliChan := make(chan *clientv3.Client, 1)

	go func() {
		cli, err := clientv3.New(config)

		if err != nil {
			errChan <- err
			return
		}

		statusRes, err := cli.Status(context.Background(), config.Endpoints[0])

		if err != nil {
			errChan <- err
			return
		} else if statusRes == nil {
			errChan <- errors.New("The status response form etcd was nil")
			return
		}

		cliChan <- cli
	}()

	select {
	case err := <-errChan:
		panic(err)
	case cli := <-cliChan:
		return &EtcdClient{
			client: cli,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to etcd cluster"))
	}
}

func (etcdCli *EtcdClient) clearAllClusterData(measureElapsedTime bool) {}
func (etcdCli *EtcdClient) clearClusterDataOneByOne() {}
func (etcdCli *EtcdClient) storeDataInCluster() {}
func (etcdCli *EtcdClient) listClusterData() {}