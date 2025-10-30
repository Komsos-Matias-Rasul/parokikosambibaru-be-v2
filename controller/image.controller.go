package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/services"
	"github.com/gin-gonic/gin"
)

var SourceGCS string = "google-cloud"
var SourceInput string = "user-input"

func (c *Controller) RenameArticleHeadline(ctx *gin.Context) {
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	type Request struct {
		NewHeadline string `json:"newHeadline"`
		Source      string `json:"source"`
	}
	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	if strings.TrimSpace(payload.NewHeadline) == "" {
		c.res.AbortInvalidRequestBody(ctx,
			errors.New("newHeadline is empty"),
			"source must be either google-cloud or user-input",
			payload,
		)
		return
	}
	validExt := regexp.MustCompile(`(?i)\.(png|jpe?g|webp)$`)
	if !validExt.MatchString(payload.NewHeadline) {
		c.res.AbortInvalidRequestBody(ctx,
			errors.New("invalid file extension"),
			`newHeadline must have extension (.png, .jpg, .jpeg, .webp)`,
			payload,
		)
		return
	}
	if payload.Source != SourceGCS && payload.Source != SourceInput {
		c.res.AbortInvalidRequestBody(ctx,
			errors.New("invalid source field"),
			"source must be either google-cloud or user-input",
			payload,
		)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	if payload.Source == SourceGCS {
		now := time.Now().UTC()
		if _, err = c.db.ExecContext(_context, `
		UPDATE articles
		SET cover_img = ?, updated_at = ?
		WHERE id = ?
		`, payload.NewHeadline, now, parsedArticleId); err != nil {
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
			"updatedAt": now,
		})
		return
	}

	var oldHeadline, oldThumbnail string
	var year int
	err = c.db.QueryRowContext(_context, `
	SELECT articles.cover_img, articles.thumb_img, editions.edition_year
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

	escapedOldHeadlineObj := strings.TrimPrefix(oldHeadline, "/")
	escapedOldThumbnailObj := strings.TrimPrefix(oldThumbnail, "/")

	// remove headline file extension
	splitNewHeadline := strings.Split(payload.NewHeadline, ".")
	trimNewHeadline := strings.TrimSuffix(
		payload.NewHeadline, fmt.Sprintf(".%s", splitNewHeadline[len(splitNewHeadline)-1]))

	// this used to store into db
	escapedNewHeadline := fmt.Sprintf("/zaitun/articles/%d/%d/%s",
		year, parsedArticleId, payload.NewHeadline)
	escapedNewThumbnail := fmt.Sprintf("/zaitun/articles/%d/%d/thumb_%s.jpg",
		year, parsedArticleId, trimNewHeadline)

	// this used to update object in GCS
	newHeadlineObj := fmt.Sprintf("zaitun/articles/%d/%d/%s",
		year, parsedArticleId, payload.NewHeadline)
	newThumbnailObj := fmt.Sprintf("zaitun/articles/%d/%d/thumb_%s.jpg",
		year, parsedArticleId, trimNewHeadline)

	// from "file%20name.jpg" to "file name.jpg"
	oldHeadlineObj, err := url.PathUnescape(escapedOldHeadlineObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode headline",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}
	oldThumbnailObj, err := url.PathUnescape(escapedOldThumbnailObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode thumbnail",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}
	newHeadlineObj, err = url.PathUnescape(newHeadlineObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode headline",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}
	newThumbnailObj, err = url.PathUnescape(newThumbnailObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode thumbnail",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}

	tx, err := c.db.BeginTx(_context, nil)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}
	if _, err = tx.ExecContext(_context, `
		UPDATE articles
		SET cover_img = ?, thumb_img = ?
		WHERE id = ?
	`, escapedNewHeadline, escapedNewThumbnail, parsedArticleId); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	_, err = services.MoveObject(_context, oldHeadlineObj, newHeadlineObj)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}
	thumbnailAttrs, err := services.MoveObject(_context, oldThumbnailObj, newThumbnailObj)
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

func (c *Controller) RenameEditionCover(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	type Request struct {
		NewCover string `json:"newCover"`
		Source   string `json:"source"`
	}
	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	if strings.TrimSpace(payload.NewCover) == "" {
		c.res.AbortInvalidRequestBody(ctx,
			errors.New("newCover is empty"),
			"source must be either google-cloud or user-input",
			payload,
		)
		return
	}
	validExt := regexp.MustCompile(`(?i)\.(png|jpe?g|webp)$`)
	if !validExt.MatchString(payload.NewCover) {
		c.res.AbortInvalidRequestBody(ctx,
			errors.New("invalid file extension"),
			`newHeadline must have extension (.png, .jpg, .jpeg, .webp)`,
			payload,
		)
		return
	}
	if payload.Source != SourceGCS && payload.Source != SourceInput {
		c.res.AbortInvalidRequestBody(ctx,
			errors.New("invalid source field"),
			"source must be either google-cloud or user-input",
			payload,
		)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	if payload.Source == SourceGCS {
		if _, err = c.db.ExecContext(_context, `
		UPDATE editions
		SET cover_img = ?
		WHERE id = ?
		`, payload.NewCover, parsedEditionId); err != nil {
			c.res.AbortDatabaseError(ctx, err, payload)
			return
		}
		if _context.Err() == context.DeadlineExceeded {
			c.res.AbortDatabaseTimeout(ctx, _context.Err(), payload)
			return
		}

		c.res.SuccessWithStatusOKJSON(ctx, payload, gin.H{
			"message": "Filename updated successfully",
			"id":      parsedEditionId,
		})
		return
	}

	var oldCover, oldThumbnail string
	var year int
	err = c.db.QueryRowContext(_context, `
	SELECT cover_img, thumb_img, edition_year
	FROM editions
	WHERE id = ?`, parsedEditionId).Scan(
		&oldCover, &oldThumbnail, &year)
	if err == sql.ErrNoRows {
		c.res.AbortEditionNotFound(ctx, err, err.Error(), payload)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	if oldCover == "/static/placeholder.jpg" || oldThumbnail == "/static/placeholder.jpg" {
		err := errors.New("cannot modify default object")
		c.res.AbortWithStatusJSON(ctx,
			err, err.Error(), "", http.StatusUnauthorized, payload)
		return
	}

	escapedOldCoverObj := strings.TrimPrefix(oldCover, "/")
	escapedOldThumbnailObj := strings.TrimPrefix(oldThumbnail, "/")

	// remove headline file extension
	splitNewCover := strings.Split(payload.NewCover, ".")
	trimNewCover := strings.TrimSuffix(
		payload.NewCover, fmt.Sprintf(".%s", splitNewCover[len(splitNewCover)-1]))

	// this used to store into db
	escapedNewCover := fmt.Sprintf("/zaitun/editions/%d/%d/%s",
		year, parsedEditionId, payload.NewCover)
	escapedNewThumbnail := fmt.Sprintf("/zaitun/editions/%d/%d/thumb_%s.jpg",
		year, parsedEditionId, trimNewCover)

	// this used to update object in GCS
	newCoverObj := fmt.Sprintf("zaitun/editions/%d/%d/%s",
		year, parsedEditionId, payload.NewCover)
	newThumbnailObj := fmt.Sprintf("zaitun/editions/%d/%d/thumb_%s.jpg",
		year, parsedEditionId, trimNewCover)

	// from "file%20name.jpg" to "file name.jpg"
	oldCoverObj, err := url.PathUnescape(escapedOldCoverObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode cover",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}
	oldThumbnailObj, err := url.PathUnescape(escapedOldThumbnailObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode thumbnail",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}
	newCoverObj, err = url.PathUnescape(newCoverObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode cover",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}
	newThumbnailObj, err = url.PathUnescape(newThumbnailObj)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to decode thumbnail",
			err.Error(), http.StatusInternalServerError, payload)
		return
	}

	tx, err := c.db.BeginTx(_context, nil)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}
	if _, err = tx.ExecContext(_context, `
		UPDATE editions
		SET cover_img = ?, thumb_img = ?
		WHERE id = ?
	`, escapedNewCover, escapedNewThumbnail, parsedEditionId); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	_, err = services.MoveObject(_context, oldCoverObj, newCoverObj)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}
	thumbnailAttrs, err := services.MoveObject(_context, oldThumbnailObj, newThumbnailObj)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}

	// if _, err = tx.ExecContext(_context, `
	// 	UPDATE articles
	// 	SET updated_at = ?
	// 	WHERE id = ?
	// `, thumbnailAttrs.Updated, parsedArticleId); err != nil {
	// 	c.res.AbortDatabaseError(ctx, err, payload)
	// 	return
	// }
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
		"id":        parsedEditionId,
		"updatedAt": thumbnailAttrs.Updated,
	})
}

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
		SELECT id, cover_img, thumb_img
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
