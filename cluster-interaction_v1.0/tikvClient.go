package main

import (
	"context"
	"errors"
	"time"

	"github.com/tikv/client-go/v2/config"
	"github.com/tikv/client-go/v2/rawkv"
)

type TikvClient struct {
	client *rawkv.Client
}

func newTikvClient() IClusterClient {
	errChan := make(chan error, 1)
	cliChan := make(chan *rawkv.Client, 1)

	go func() {
		cli, err := rawkv.NewClient(context.Background(), []string{"127.0.0.1:2379"}, config.DefaultConfig().Security)

		if err != nil {
			errChan <- err
			return
		}

		cliChan <- cli
	}()

	select {
	case err := <-errChan:
		panic(err)
	case cli := <-cliChan:
		return &TikvClient{
			client: cli,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to TiKV cluster"))
	}
}

func (tikvCli *TikvClient) clearAllClusterData(measureElapsedTime bool) {}
func (tikvCli *TikvClient) clearClusterDataOneByOne() {}
func (tikvCli *TikvClient) storeDataInCluster() {}
func (tikvCli *TikvClient) listClusterData() {}