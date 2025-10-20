package controller

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"time"

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
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	type Cover struct {
		ArticleId    *int    `json:"articleId"`
		CoverImg     *string `json:"coverImg"`
		ThumbnailImg *string `json:"thumbnailImg"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var cover Cover
	err = c.db.QueryRowContext(_context, `
		SELECT id, headline_img, thumb_img
		FROM articles
    WHERE id = ?`, parsedArticleId).Scan(
		&cover.ArticleId,
		&cover.CoverImg,
		&cover.ThumbnailImg,
	)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err == sql.ErrNoRows {
		c.res.AbortArticleNotFound(ctx, err, "", nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, cover)
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
