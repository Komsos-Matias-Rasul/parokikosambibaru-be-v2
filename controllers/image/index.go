package image

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

var SourceGCS string = "google-cloud"
var SourceInput string = "user-input"

type ImageController struct {
	db  *sql.DB
	res *lib.Responses
}

func NewImageController(db *sql.DB, res *lib.Responses) *ImageController {
	return &ImageController{db, res}
}

func ValidateImgRequests(fileName string, contentType string) error {
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

func (c *ImageController) ViewBucketAttributes(ctx *gin.Context) {
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
