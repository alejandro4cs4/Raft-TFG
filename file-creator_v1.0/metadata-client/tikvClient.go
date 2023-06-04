package metadataclient

import (
	"context"
	"errors"
	"time"

	"github.com/tikv/client-go/v2/config"
	"github.com/tikv/client-go/v2/rawkv"

	"Raft-TFG/file-creator_v1.0/settings"
)

type TikvClient struct {
	client   *rawkv.Client
	settings *settings.Settings
}

func newTikvClient(settings *settings.Settings) IMetadataClient {
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
			client:   cli,
			settings: settings,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to TiKV metadata storage"))
	}
}

func (tikvCli *TikvClient) CloseClient() {
	tikvCli.client.Close()
}

func (tikvCli *TikvClient) StoreKeyValue(key string, value string) {

}

func (tikvCli *TikvClient) GetByKey(key string) MetadataClientGetResponse {
	return MetadataClientGetResponse{
		Count: 0,
	}
}
