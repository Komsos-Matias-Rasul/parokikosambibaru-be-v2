package controller

import (
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) GetZaitunCoverImg(ctx *gin.Context) {
	year := ctx.Param("year")
	editionId := ctx.Param("editionId")
	fileName := ctx.Param("fileName")

	client, err := lib.GetCloudStorage(ctx.Request.Context())
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}
	defer client.CloudStorageClient.Close()

	_fileName := fmt.Sprintf("zaitun/editions/%s/%s/%s", year, editionId, fileName)
	obj := client.StorageBucket.Object(_fileName)

	reader, err := obj.NewReader(ctx.Request.Context())
	if err == storage.ErrObjectNotExist {
		c.res.AbortNoObject(ctx, err, nil)
		return
	}
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}

	p, err := io.ReadAll(reader)
	if err != nil {
		c.res.AbortReadFailure(ctx, err, nil)
		return
	}
	defer reader.Close()

	attr, _ := obj.Attrs(ctx.Request.Context())
	c.res.SuccessWithData(ctx, attr.ContentType, []byte(p), fileName)
}

func (c *Controller) GetArticleCoverImg(ctx *gin.Context) {
	year := ctx.Param("year")
	articleId := ctx.Param("articleId")
	fileName := ctx.Param("fileName")

	client, err := lib.GetCloudStorage(ctx.Request.Context())
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}
	defer client.CloudStorageClient.Close()

	_fileName := fmt.Sprintf("zaitun/articles/%s/%s/%s", year, articleId, fileName)
	obj := client.StorageBucket.Object(_fileName)
	reader, err := obj.NewReader(ctx.Request.Context())
	if err == storage.ErrObjectNotExist {
		c.res.AbortNoObject(ctx, err, nil)
		return
	}
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}

	p, err := io.ReadAll(reader)
	if err != nil {
		c.res.AbortReadFailure(ctx, err, nil)
		return
	}
	defer reader.Close()

	attr, _ := obj.Attrs(ctx.Request.Context())
	c.res.SuccessWithData(ctx, attr.ContentType, []byte(p), fileName)
}

func (c *Controller) GetAdImage(ctx *gin.Context) {
	year := ctx.Param("year")
	fileName := ctx.Param("fileName")

	client, err := lib.GetCloudStorage(ctx.Request.Context())
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}
	defer client.CloudStorageClient.Close()

	_fileName := fmt.Sprintf("ads/%s/%s", year, fileName)
	obj := client.StorageBucket.Object(_fileName)
	reader, err := obj.NewReader(ctx.Request.Context())
	if err == storage.ErrObjectNotExist {
		c.res.AbortNoObject(ctx, err, nil)
		return
	}
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}

	p, err := io.ReadAll(reader)
	if err != nil {
		c.res.AbortReadFailure(ctx, err, nil)
		return
	}
	defer reader.Close()

	attr, _ := obj.Attrs(ctx.Request.Context())
	c.res.SuccessWithData(ctx, attr.ContentType, []byte(p), fileName)
}
