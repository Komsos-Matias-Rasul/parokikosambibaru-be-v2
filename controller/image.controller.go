package controller

import (
	"fmt"
	"io"
	"net/http"

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
		res := gin.H{"error": lib.ErrStorage.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer client.CloudStorageClient.Close()

	_fileName := fmt.Sprintf("zaitun/editions/%s/%s/%s", year, editionId, fileName)
	obj := client.StorageBucket.Object(_fileName)

	reader, err := obj.NewReader(ctx.Request.Context())
	if err == storage.ErrObjectNotExist {
		res := gin.H{"error": lib.ErrNoObject.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if err != nil {
		res := gin.H{"error": lib.ErrStorage.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	p, err := io.ReadAll(reader)
	if err != nil {
		res := gin.H{"error": lib.ErrReadFailure.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer reader.Close()

	attr, _ := obj.Attrs(ctx.Request.Context())
	ctx.Data(http.StatusOK, attr.ContentType, []byte(p))
	c.logger.Info(ctx.Copy(), nil, fileName)
}

func (c *Controller) GetArticleCoverImg(ctx *gin.Context) {
	year := ctx.Param("year")
	articleId := ctx.Param("articleId")
	fileName := ctx.Param("fileName")

	client, err := lib.GetCloudStorage(ctx.Request.Context())
	if err != nil {
		res := gin.H{"error": lib.ErrStorage.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer client.CloudStorageClient.Close()

	_fileName := fmt.Sprintf("zaitun/articles/%s/%s/%s", year, articleId, fileName)
	obj := client.StorageBucket.Object(_fileName)
	reader, err := obj.NewReader(ctx.Request.Context())
	if err == storage.ErrObjectNotExist {
		res := gin.H{"error": lib.ErrNoObject.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if err != nil {
		res := gin.H{"error": lib.ErrStorage.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	p, err := io.ReadAll(reader)
	if err != nil {
		res := gin.H{"error": lib.ErrReadFailure.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer reader.Close()

	attr, _ := obj.Attrs(ctx.Request.Context())
	ctx.Data(http.StatusOK, attr.ContentType, []byte(p))
	c.logger.Info(ctx.Copy(), nil, fileName)
}

func (c *Controller) GetAdImage(ctx *gin.Context) {
	year := ctx.Param("year")
	fileName := ctx.Param("fileName")

	client, err := lib.GetCloudStorage(ctx.Request.Context())
	if err != nil {
		res := gin.H{"error": lib.ErrStorage.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer client.CloudStorageClient.Close()

	_fileName := fmt.Sprintf("ads/%s/%s", year, fileName)
	obj := client.StorageBucket.Object(_fileName)
	reader, err := obj.NewReader(ctx.Request.Context())
	if err == storage.ErrObjectNotExist {
		res := gin.H{"error": lib.ErrNoObject.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if err != nil {
		res := gin.H{"error": lib.ErrStorage.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	p, err := io.ReadAll(reader)
	if err != nil {
		res := gin.H{"error": lib.ErrReadFailure.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer reader.Close()

	attr, _ := obj.Attrs(ctx.Request.Context())
	ctx.Data(http.StatusOK, attr.ContentType, []byte(p))
	c.logger.Info(ctx.Copy(), nil, fileName)
}
