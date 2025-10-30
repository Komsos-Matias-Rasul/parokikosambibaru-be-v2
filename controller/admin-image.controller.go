package controller

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/services"
	"github.com/gin-gonic/gin"
)

func validateImgRequests(fileName string, contentType string) error {
	if strings.TrimSpace(fileName) == "" {
		return errors.New("empty filename")
	}
	invalidChars := regexp.MustCompile(`[\/\\?%*:|"<>^]`)
	if invalidChars.MatchString(fileName) {
		return errors.New("invalid characters")
	}
	validExt := regexp.MustCompile(`(?i)\.(png|jpe?g|webp)$`)
	if !validExt.MatchString(fileName) {
		return errors.New("invalid file extension")
	}
	if strings.TrimSpace(contentType) == "" {
		return errors.New("empty content type")
	}
	validType := regexp.MustCompile(`(?i)image\/(png|jpe?g|webp)$`)
	if !validType.MatchString(contentType) {
		return errors.New("invalid content type")
	}
	return nil
}

func (c *Controller) CoreSaveEditionCover(ctx *gin.Context) {
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
	if err := validateImgRequests(payload.FileName, payload.ContentType); err != nil {
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

func (c *Controller) CoreSaveArticleCover(ctx *gin.Context) {
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
	if err := validateImgRequests(payload.FileName, payload.ContentType); err != nil {
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

func (c *Controller) CoreUpdateArticleThumbnail(ctx *gin.Context) {
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

func (c *Controller) CoreUpdateEditionThumbnail(ctx *gin.Context) {
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

func (c *Controller) CoreSaveArticleImageContents(ctx *gin.Context) {
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
	if err := validateImgRequests(payload.FileName, payload.ContentType); err != nil {
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

func (c *Controller) ViewBucketAttributes(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	s, err := lib.GetCloudStorage(_context)
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}
	defer s.CloudStorageClient.Close()

	attrs, err := s.StorageBucket.Attrs(_context)
	if err != nil {
		c.res.AbortStorageError(ctx, err, nil)
		return
	}
	c.res.SuccessWithStatusOKJSON(ctx, nil, attrs)
}
