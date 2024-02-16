package minio

import (
	"io"
	"log"
	"os"

	"github.com/minio/minio-go"
)

func CreateMinioClient() *minio.Client {
	endpoint := "localhost:9000"
	if val, ok := os.LookupEnv("MINIO_URL"); ok {
		endpoint = val
	}
	accessKeyID := "minioadmin"
	if val, ok := os.LookupEnv("MINIO_ROOT_USER"); ok {
		accessKeyID = val
	}
	secretAccessKey := "minioadmin"
	if val, ok := os.LookupEnv("MINIO_ROOT_PASSWORD"); ok {
		secretAccessKey = val
	}
	useSSL := false
	if val, ok := os.LookupEnv("MINIO_USE_SSL"); ok {
		useSSL = val == "true"
	}

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalf("Failed to create minio client: %v", err)
	}

	return minioClient
}

func CreateBucket(minioClient *minio.Client, bucketName string, location string) {
	err := minioClient.MakeBucket(bucketName, location)
	if err != nil {
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("Bucket already exists: %s\n", bucketName)
		} else {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	} else {
		log.Printf("Successfully created bucket: %s\n", bucketName)
	}
}

func UploadFile(minioClient *minio.Client, bucketName string, objectName string, reader io.Reader, objectSize int64, contentType string) {
	_, err := minioClient.PutObject(bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
}

func DownloadFile(minioClient *minio.Client, bucketName string, objectName string) *minio.Object {
	object, err := minioClient.GetObject(bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	return object
}
