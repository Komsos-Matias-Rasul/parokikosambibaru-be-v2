package image

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/services"
	"github.com/gin-gonic/gin"
)

func (c *ImageController) GetArticleCoverImg(ctx *gin.Context) {
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

func (c *ImageController) SaveArticleCover(ctx *gin.Context) {
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
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
		JOIN articles ON editions.id = articles.edition_id
		WHERE articles.id = ?`, parsedArticleId).Scan(&year); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	obj := fmt.Sprintf("zaitun/articles/%d/%d/%s", year, parsedArticleId, payload.FileName)
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

func (c *ImageController) UpdateArticleThumbnail(ctx *gin.Context) {
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
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
		UPDATE articles
		SET thumb_img = ?
		WHERE id = ?`, payload.FileName, parsedArticleId); err != nil {
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

func (c *ImageController) SaveArticleImageContents(ctx *gin.Context) {
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
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
		JOIN articles ON editions.id = articles.edition_id
		WHERE articles.id = ?`, parsedArticleId).Scan(&year); err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	obj := fmt.Sprintf("zaitun/articles/%d/%d/%s", year, parsedArticleId, payload.FileName)
	signedUrl, err := services.GetSignedURL(_context, obj, payload.ContentType)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), payload)
		return
	}

	// remove https://storage.googleapis.com/<bucket> prefix
	_, loc, _ := strings.Cut(signedUrl, conf.GCLOUD_BUCKET)

	// remove url queries
	loc, _, _ = strings.Cut(loc, "?")

	c.res.SuccessWithStatusOKJSON(ctx, payload, gin.H{"url": signedUrl, "location": loc})
}

func (c *ImageController) RenameArticleHeadline(ctx *gin.Context) {
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
