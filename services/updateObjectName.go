package services

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

func MoveObject(ctx context.Context, oldPath string, destination string) (*storage.ObjectAttrs, error) {
	client, err := lib.GetCloudStorage(ctx)
	if err != nil {
		return nil, err
	}
	defer client.CloudStorageClient.Close()

	w := client.StorageBucket.Object(oldPath)
	attr, err := w.Move(ctx, storage.MoveObjectDestination{
		Object: destination,
	})
	if err != nil {
		return nil, err
	}

	return attr, nil
}
