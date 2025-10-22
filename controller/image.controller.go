package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/services"
	"github.com/gin-gonic/gin"
)

func (c *Controller) UpdateArticleImgFileName(ctx *gin.Context) {
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	type Request struct {
		NewHeadline   string `json:"newHeadline"`
		FileExtension string `json:"fileExtension"`
	}
	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var oldHeadline, oldThumbnail string
	var year int
	err = c.db.QueryRowContext(_context, `
	SELECT articles.headline_img, articles.thumb_img, editions.edition_year
	FROM articles
	JOIN editions ON articles.edition_id = editions.id
	WHERE articles.id = ?`, parsedArticleId).Scan(
		&oldHeadline, &oldThumbnail, &year)
	if err == sql.ErrNoRows {
		c.res.AbortArticleNotFound(ctx, err, err.Error(), payload)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	if oldHeadline == "/static/placeholder.jpg" || oldThumbnail == "/static/placeholder.jpg" {
		err := errors.New("cannot modify default object")
		c.res.AbortWithStatusJSON(ctx,
			err, err.Error(), "", http.StatusUnauthorized, payload)
		return
	}

	oldHeadline, _ = strings.CutPrefix(oldHeadline, "/")
	oldThumbnail, _ = strings.CutPrefix(oldThumbnail, "/")

	newHeadline := fmt.Sprintf("zaitun/articles/%d/%d/%s.%s",
		year, parsedArticleId, payload.NewHeadline, payload.FileExtension)
	newThumbnail := fmt.Sprintf("zaitun/articles/%d/%d/%s_thumb.webp",
		year, parsedArticleId, payload.NewHeadline)

	tx, err := c.db.BeginTx(_context, nil)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}
	if _, err = tx.ExecContext(_context, `
		UPDATE articles
		SET headline_img = ?, thumb_img = ?
		WHERE id = ?
	`, fmt.Sprintf("/%s", newHeadline), fmt.Sprintf("/%s", newThumbnail), parsedArticleId); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	_, err = services.MoveObject(_context, oldHeadline, newHeadline)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}
	thumbnailAttrs, err := services.MoveObject(_context, oldThumbnail, newThumbnail)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}

	if _, err = tx.ExecContext(_context, `
		UPDATE articles
		SET updated_at = ?
		WHERE id = ?
	`, thumbnailAttrs.Updated, parsedArticleId); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}
	if err := tx.Commit(); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), payload)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, payload, gin.H{
		"message":   "Filename updated successfully",
		"id":        parsedArticleId,
		"updatedAt": thumbnailAttrs.Updated,
	})
}

// func (c *Controller) UpdateEditionImgFileName(ctx *gin.Context) {
// 	editionId := ctx.Param("editionId")
// 	parsedEditionId, err := strconv.Atoi(editionId)
// 	if err != nil {
// 		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
// 		return
// 	}

// 	type Request struct {
// 		NewHeadline   string `json:"newHeadline"`
// 		FileExtension string `json:"fileExtension"`
// 	}
// 	var payload Request
// 	if err := ctx.BindJSON(&payload); err != nil {
// 		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
// 		return
// 	}

// 	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
// 	defer cancel()

// 	var oldHeadline, oldThumbnail string
// 	var year int
// 	err = c.db.QueryRowContext(_context, `
// 	SELECT articles.headline_img, articles.thumb_img, editions.edition_year
// 	FROM articles
// 	JOIN editions ON articles.edition_id = editions.id
// 	WHERE articles.id = ?`, parsedArticleId).Scan(
// 		&oldHeadline, &oldThumbnail, &year)
// 	if err == sql.ErrNoRows {
// 		c.res.AbortArticleNotFound(ctx, err, err.Error(), payload)
// 		return
// 	}
// 	if err != nil {
// 		c.res.AbortDatabaseError(ctx, err, payload)
// 		return
// 	}

// 	oldHeadline, _ = strings.CutPrefix(oldHeadline, "/")
// 	oldThumbnail, _ = strings.CutPrefix(oldThumbnail, "/")

// 	newHeadline := fmt.Sprintf("zaitun/articles/%d/%d/%s.%s",
// 		year, parsedArticleId, payload.NewHeadline, payload.FileExtension)
// 	newThumbnail := fmt.Sprintf("zaitun/articles/%d/%d/%s_thumb.webp",
// 		year, parsedArticleId, payload.NewHeadline)

// 	tx, err := c.db.BeginTx(_context, nil)
// 	if err != nil {
// 		c.res.AbortDatabaseError(ctx, err, payload)
// 		return
// 	}
// 	if _, err = tx.ExecContext(_context, `
// 		UPDATE articles
// 		SET headline_img = ?, thumb_img = ?
// 		WHERE id = ?
// 	`, fmt.Sprintf("/%s", newHeadline), fmt.Sprintf("/%s", newThumbnail), parsedArticleId); err != nil {
// 		c.res.AbortDatabaseError(ctx, err, payload)
// 		return
// 	}

// 	_, err = services.MoveObject(_context, oldHeadline, newHeadline)
// 	if err != nil {
// 		c.res.AbortStorageError(ctx, err, payload)
// 		return
// 	}
// 	thumbnailAttrs, err := services.MoveObject(_context, oldThumbnail, newThumbnail)
// 	if err != nil {
// 		c.res.AbortStorageError(ctx, err, payload)
// 		return
// 	}

// 	if _, err = tx.ExecContext(_context, `
// 		UPDATE articles
// 		SET updated_at = ?
// 		WHERE id = ?
// 	`, thumbnailAttrs.Updated, parsedArticleId); err != nil {
// 		c.res.AbortDatabaseError(ctx, err, payload)
// 		return
// 	}
// 	if err := tx.Commit(); err != nil {
// 		c.res.AbortDatabaseError(ctx, err, payload)
// 		return
// 	}

// 	if _context.Err() == context.DeadlineExceeded {
// 		c.res.AbortDatabaseTimeout(ctx, _context.Err(), payload)
// 		return
// 	}

// 	c.res.SuccessWithStatusOKJSON(ctx, payload, gin.H{
// 		"message":   "Filename updated successfully",
// 		"id":        parsedArticleId,
// 		"updatedAt": thumbnailAttrs.Updated,
// 	})
// }

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
