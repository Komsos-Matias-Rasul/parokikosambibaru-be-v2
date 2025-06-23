package lib

import (
	"context"
	"encoding/base64"
	"log"

	"cloud.google.com/go/storage"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"google.golang.org/api/option"
)

type CloudStorage struct {
	CloudStorageClient *storage.Client
	StorageBucket      *storage.BucketHandle
}

func GetCloudStorage(ctx context.Context) (*CloudStorage, error) {
	b64, err := base64.RawStdEncoding.DecodeString(conf.GOOGLE_CREDENTIALS_BASE64)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, err
	}

	opt := option.WithCredentialsJSON(b64)
	s, err := storage.NewClient(ctx, opt)
	if err != nil {
		log.Default().Println(err.Error())
		return nil, err
	}

	bkt := s.Bucket(conf.GCLOUD_BUCKET)
	return &CloudStorage{
		CloudStorageClient: s,
		StorageBucket:      bkt,
	}, nil
}
