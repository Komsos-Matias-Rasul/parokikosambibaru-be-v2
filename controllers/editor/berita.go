package editor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/services"
	"github.com/gin-gonic/gin"
)

func (c *EditorController) GetAllBerita(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := c.db.QueryContext(_context, `
	SELECT id, title, section, thumb_img, descriptions,
	details, created_at, publish_start, publish_end FROM announcements
		WHERE deleted_at is null ORDER BY publish_start DESC`)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()

	type Berita struct {
		Id           int        `json:"id"`
		Title        string     `json:"title"`
		Section      string     `json:"section"`
		ThumbImg     string     `json:"thumbImg"`
		Desc         string     `json:"descriptions"`
		Details      string     `json:"details"`
		CreatedAt    *time.Time `json:"createdAt"`
		PublishStart *time.Time `json:"publishStart"`
		PublishEnd   *time.Time `json:"publishEnd"`
	}

	news := []*Berita{}
	for rows.Next() {
		var result Berita
		var createdAt, publishStart, publishEnd []uint8
		if err := rows.Scan(
			&result.Id,
			&result.Title,
			&result.Section,
			&result.ThumbImg,
			&result.Desc,
			&result.Details,
			&createdAt,
			&publishStart,
			&publishEnd,
		); err != nil {
			log.Println(err)
		}

		result.CreatedAt = lib.Base64ToTime(createdAt)
		result.PublishStart = lib.Base64ToTime(publishStart)
		result.PublishEnd = lib.Base64ToTime(publishEnd)
		news = append(news, &result)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, news)
}

func (c *EditorController) CreateBerita(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	type RequestModel struct {
		Title        string `json:"title" binding:"required"`
		Section      string `json:"section" binding:"required"`
		Desc         string `json:"descriptions" binding:"required"`
		Details      string `json:"details" binding:"required"`
		PublishStart string `json:"publishStart" binding:"required"`
		PublishEnd   string `json:"publishEnd" binding:"required"`
		FileName     string `json:"thumbImg" binding:"required"`
		ContentType  string `json:"contentType" binding:"required"`
	}

	var payload RequestModel
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	pubStart, err := time.Parse(time.RFC3339, payload.PublishStart)
	if err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	pubEnd, err := time.Parse(time.RFC3339, payload.PublishEnd)
	if err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	imgPath := "/static/placeholder.jpg"
	result, err := c.db.ExecContext(_context, `
	INSERT INTO announcements
	(title, section, thumb_img, descriptions, details, publish_start, publish_end)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		payload.Title,
		payload.Section,
		imgPath,
		payload.Desc,
		payload.Details,
		pubStart,
		pubEnd,
	)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	beritaId, err := result.LastInsertId()
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	obj := fmt.Sprintf("berita/%d/%s", beritaId, payload.FileName)

	signedUrl, err := services.GetSignedURL(_context, obj, payload.ContentType)
	if err != nil {
		c.res.AbortStorageError(ctx, err, payload)
		return
	}
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), payload)
		return
	}
	res := gin.H{"message": "berita created successfully", "url": signedUrl, "location": obj, "beritaId": beritaId}
	c.res.SuccessWithStatusJSON(ctx, http.StatusCreated, nil, res)
}

func (c *EditorController) UpdateBeritaThumbnail(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	id := ctx.Param("id")
	parsedBeritaId, err := strconv.Atoi(id)
	if err != nil {
		c.res.AbortInvalidBerita(ctx, err, err.Error(), nil)
		return
	}

	type Request struct {
		ObjPath string `json:"objPath"`
	}
	var payload Request
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}
	if strings.TrimSpace(payload.ObjPath) == "" {
		c.res.AbortInvalidRequestBody(
			ctx,
			errors.New("empty object path"),
			"empty object path",
			nil,
		)
		return
	}

	type CompressRequest struct {
		Path          string `json:"path"`
		BucketName    string `json:"bucketName"`
		TargetType    string `json:"targetType"`
		TargetQuality int    `json:"targetQuality"`
	}

	body := CompressRequest{
		Path:          payload.ObjPath,
		BucketName:    conf.GCLOUD_BUCKET,
		TargetType:    "webp",
		TargetQuality: 80,
	}
	bBody, err := json.Marshal(body)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to call image compression API",
			err.Error(), http.StatusInternalServerError, body,
		)
		return
	}
	bodyReader := bytes.NewReader(bBody)
	resp, err := http.Post(conf.COMPRESS_FUNCTION_URL, "application/json", bodyReader)
	if err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to call image compression API",
			err.Error(), http.StatusInternalServerError, body,
		)
		return
	}
	defer resp.Body.Close()

	type CompressResponse struct {
		Url string `json:"url"`
	}
	var comp CompressResponse
	if err = json.NewDecoder(resp.Body).Decode(&comp); err != nil {
		c.res.AbortWithStatusJSON(
			ctx, err, "failed to read compress API response",
			err.Error(), http.StatusInternalServerError, body,
		)
		return
	}

	if _, err := c.db.ExecContext(_context, `
		UPDATE announcements
		SET thumb_img = ?
		WHERE id = ?`, comp.Url, parsedBeritaId); err != nil {
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

func (c *EditorController) UpdateBeritaPublishing(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()
	beritaId := ctx.Param("id")
	if beritaId == "" {
		c.res.AbortInvalidRequestBody(ctx, nil, "berita id is required", nil)
		return
	}

	type RequestModel struct {
		PublishStart string `json:"publishStart" binding:"required"`
		PublishEnd   string `json:"publishEnd" binding:"required"`
	}

	var payload RequestModel
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	pubStart, err := time.Parse(time.RFC3339, payload.PublishStart)
	if err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, "Invalid publishStart format (use ISO 8601)", nil)
		return
	}
	pubEnd, err := time.Parse(time.RFC3339, payload.PublishEnd)
	if err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, "Invalid publishEnd format (use ISO 8601)", nil)
		return
	}

	query := `UPDATE announcements SET publish_start = ?, publish_end = ? WHERE id = ?`

	result, err := c.db.ExecContext(_context, query, pubStart, pubEnd, beritaId)

	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.res.AbortDatabaseError(ctx, nil, "Berita not found")
		return
	}

	responsePayload := gin.H{
		"_id": beritaId,
		"data": gin.H{
			"message": "publishing period updated successfully",
		},
		"timestamp": time.Now().UTC().Unix(),
	}

	c.res.SuccessWithStatusJSON(ctx, http.StatusOK, nil, responsePayload)
}

func (c *EditorController) DeleteBeritaPermanent(ctx *gin.Context) {
	beritaid := ctx.Param("id")
	id, err := strconv.Atoi(beritaid)
	if err != nil {
		c.res.AbortInvalidBerita(ctx, err, err.Error(), nil)
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM announcements WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	if !exists {
		c.res.AbortArticleNotFound(ctx, err, "", nil)
		return
	}

	_, err = c.db.Exec(`
		DELETE FROM announcements
		WHERE id = ?`,
		id,
	)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	res := gin.H{"message": "berita deleted successfully"}
	c.res.SuccessWithStatusJSON(ctx, http.StatusAccepted, nil, res)
}
