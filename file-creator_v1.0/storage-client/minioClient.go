package storageclient

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"Raft-TFG/file-creator_v1.0/settings"
)

const (
	testBucket string = "testbucket"
)

type MinioClient struct {
	client   *minio.Client
	settings *settings.Settings
}

func newMinioclient(settings *settings.Settings) IStorageClient {
	endpoint := "127.0.0.1:9000"
	accessKeyID := "b24ih9BtvgC3PXNA"
	secretAccessKey := "PgsCwXjqbx3J4kydzUrmkYlymaSLN8W1"
	useSSL := false

	config := minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	}

	errChan := make(chan error, 1)
	cliChan := make(chan *minio.Client, 1)
	ticker := time.NewTicker(time.Second)
	quit := make(chan struct{})

	go func() {
		cli, err := minio.New(endpoint, &config)

		if err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case <-ticker.C:
				_, err := cli.GetBucketLocation(context.Background(), "testbucket")
				if err == nil {
					cliChan <- cli
					close(quit)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		panic(err)
	case cli := <-cliChan:
		return &MinioClient{
			client:   cli,
			settings: settings,
		}
	case <-time.After(6 * time.Second):
		panic(errors.New("5 seconds timeout exceeded while trying to connect to MinIO object storage"))
	}
}

func (minioCli *MinioClient) StoreObject(objectName string, filePath string) {
	ctx := context.Background()

	_, err := minioCli.client.FPutObject(ctx, testBucket, objectName, filePath, minio.PutObjectOptions{})

	if err != nil {
		panic(err)
	}
}

func createBucket(minioClient *minio.Client, bucketName string) {
	ctx := context.Background()
	location := "us-east-1"

	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s bucket\n", bucketName)
	}
}
