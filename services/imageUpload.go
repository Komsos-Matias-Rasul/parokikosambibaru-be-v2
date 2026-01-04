package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

func GetSignedURL(ctx context.Context, destination string, contentType string) (string, error) {
	client, err := lib.GetCloudStorage(ctx)
	if err != nil {
		return "", err
	}
	defer client.CloudStorageClient.Close()

	opt := &storage.SignedURLOptions{
		Scheme: storage.SigningSchemeV4,
		Method: http.MethodPut,
		Headers: []string{
			fmt.Sprintf("Content-Type:%s", contentType),
		},
		Expires: time.Now().Add(3 * time.Minute),
	}
	signedUrl, err := client.StorageBucket.SignedURL(destination, opt)
	if err != nil {
		return "", err
	}
	return signedUrl, nil
}

func RefreshCORS(ctx context.Context, bkt *storage.BucketHandle) error {
	_attr := storage.BucketAttrsToUpdate{
		CORS: []storage.CORS{{
			Methods:         []string{http.MethodOptions, http.MethodPut},
			Origins:         []string{"*"},
			ResponseHeaders: []string{"Content-Type", "Accept"},
		}},
	}
	if _, err := bkt.Update(ctx, _attr); err != nil {
		return err
	}
	return nil
}
