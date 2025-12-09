package editor

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
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
		Desc         string `json:"description" binding:"required"`
		Details      string `json:"details" binding:"required"`
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

	res := gin.H{"message": "article created successfully", "berita_id": beritaId}
	c.res.SuccessWithStatusJSON(ctx, http.StatusCreated, nil, res)
}
