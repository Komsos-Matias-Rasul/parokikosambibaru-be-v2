package image

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

func (c *ImageController) GetEditionCoverImg(ctx *gin.Context) {
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

func (c *ImageController) SaveEditionCover(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	type Request struct {
		FileName    string `json:"fileName"`
		ContentType string `json:"contentType"`
	}
	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	if err := ValidateImgRequests(payload.FileName, payload.ContentType); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), payload)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var year int
	if err := c.db.QueryRowContext(_context, `
		SELECT edition_year FROM editions
		WHERE id = ?`, parsedEditionId).Scan(&year); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	obj := fmt.Sprintf("zaitun/editions/%d/%d/%s", year, parsedEditionId, payload.FileName)
	signedUrl, err := services.GetSignedURL(_context, obj, payload.ContentType)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), payload)
		return
	}
	c.res.SuccessWithStatusOKJSON(ctx, payload, gin.H{"url": signedUrl, "location": obj})
}

func (c *ImageController) UpdateEditionThumbnail(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	type Request struct {
		FileName string `json:"fileName"`
	}
	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	if strings.TrimSpace(payload.FileName) == "" {
		c.res.AbortInvalidRequestBody(
			ctx,
			errors.New("empty filename"),
			"empty filename",
			nil,
		)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()
	if _, err := c.db.ExecContext(_context, `
		UPDATE editions
		SET thumb_img = ?
		WHERE id = ?`, payload.FileName, parsedEditionId); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, payload)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, payload, gin.H{
		"message": "thumbnail updated successfully",
	})
}

func (c *ImageController) RenameEditionCover(ctx *gin.Context) {
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
