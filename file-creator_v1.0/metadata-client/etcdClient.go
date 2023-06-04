package metadataclient

import (
	"context"
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"Raft-TFG/file-creator_v1.0/settings"
)

type EtcdClient struct {
	client   *clientv3.Client
	settings *settings.Settings
}

func newEtcdClient(settings *settings.Settings) IMetadataClient {
	config := clientv3.Config{
		Endpoints:   []string{"localhost:23790"},
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
			client:   cli,
			settings: settings,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to etcd metadata storage"))
	}
}

func (etcdCli *EtcdClient) CloseClient() {
	etcdCli.client.Close()
}

func (etcdCli *EtcdClient) StoreKeyValue(key string, value string) {
	_, err := etcdCli.client.Put(context.Background(), key, value)

	if err != nil {
		panic(err)
	}
}

func (etcdCli *EtcdClient) GetByKey(key string) MetadataClientGetResponse {
	response, err := etcdCli.client.Get(context.Background(), key)

	checkError(err)

	if response.Count == 0 {
		return MetadataClientGetResponse{
			Count: response.Count,
		}
	}

	return MetadataClientGetResponse{
		Count: response.Count,
		Key:   string(response.Kvs[0].Key),
		Value: string(response.Kvs[0].Value),
	}
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
