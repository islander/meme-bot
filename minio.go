package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// object storage struct
type Storage struct {
	MinIO  *minio.Client
	bucket string
}

// initialize object storage
func NewStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool, bucket string) (error, Storage) {
	// Initialize minio client object.
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Println("[storage] ERROR ", err)
		return errors.New(fmt.Sprintf("Could not connect to MinIO server %s", endpoint)), Storage{}
	}
	storage := Storage{MinIO: client, bucket: bucket}
	return nil, storage
}

// create MinIO bucket to save images
func (st Storage) CreateBucket() error {
	ctx := context.Background()

	// create bucket
	err := st.MinIO.MakeBucket(ctx, st.bucket, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket
		exists, errBucketExists := st.MinIO.BucketExists(ctx, st.bucket)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", st.bucket)
		} else {
			fmt.Printf("%v\n", err)
			return errors.New(fmt.Sprintf("Could not create bucket %s", st.bucket))
		}
	} else {
		log.Printf("Successfully created %s\n", st.bucket)
	}
	return nil
}

func (st Storage) SaveImage(imageName string, imageBody []byte) error {
	ctx := context.Background()

	reader := bytes.NewReader(imageBody)
	opts := minio.PutObjectOptions{ContentType: "image/jpeg"}
	_, err := st.MinIO.PutObject(ctx, st.bucket, imageName, reader, int64(len(imageBody)), opts)

	if err != nil {
		return err
	}

	return nil
}

func (st Storage) FindImage(imageName string) (error, []byte) {
	ctx := context.Background()

	opts := minio.StatObjectOptions{}
	info, err := st.MinIO.StatObject(ctx, st.bucket, imageName, opts)

	if err != nil {
		return err, nil
	}

	size := info.Size
	log.Printf("[minio] Object size: %d\n", size)

	opts = minio.GetObjectOptions{}
	obj, err := st.MinIO.GetObject(ctx, st.bucket, imageName, opts)
	defer obj.Close()

	buffer := make([]byte, info.Size)
	obj.Read(buffer)

	log.Printf("[minio] Retrieved Object: %+v\n", obj)

	if err != nil {
		return err, nil
	}

	return nil, buffer
}
