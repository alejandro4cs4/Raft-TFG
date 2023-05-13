package main

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func getMinioClient() (minioClient *minio.Client, err error) {
	endpoint := "127.0.0.1:9000"
	accessKeyID := "b24ih9BtvgC3PXNA"
	secretAccessKey := "PgsCwXjqbx3J4kydzUrmkYlymaSLN8W1"
	useSSL := false

	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	return
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

func uploadObject(minioClient *minio.Client, bucketName, objectName, filePath, contentType string) {
	ctx := context.Background()

	info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
}

func main() {
	bucketName := "testbucket"
	objectName := "test.txt"
	filePath := "test-file.txt"
	contentType := "text/plain"

	minioClient, err := getMinioClient()

	if err != nil {
		log.Fatalln(err)
	}

	createBucket(minioClient, bucketName)

	uploadObject(minioClient, bucketName, objectName, filePath, contentType)
}
