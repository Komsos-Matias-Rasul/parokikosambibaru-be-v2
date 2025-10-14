package services

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
)

func UploadImage(file multipart.File, fileName string, ctx context.Context) error {
	client, err := lib.GetCloudStorage(ctx)
	if err != nil {
		return err
	}
	defer client.CloudStorageClient.Close()
	w := client.StorageBucket.Object(fileName).NewWriter(ctx)
	if _, err = io.Copy(w, file); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
