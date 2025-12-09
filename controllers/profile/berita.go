package profile

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *ProfileController) GetOverviewBerita(ctx *gin.Context) {

}

func (c *ProfileController) GetAllBerita(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := c.db.QueryContext(_context, `
	SELECT id, title, section, thumb_img, descriptions FROM announcements
		WHERE deleted_at is null AND publish_start <= now() AND publish_end >= now()
		ORDER BY publish_start DESC`)
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
		Id       int    `json:"id"`
		Title    string `json:"title"`
		Section  string `json:"section"`
		ThumbImg string `json:"thumbImg"`
		Desc     string `json:"descriptions"`
	}

	news := []*Berita{}
	for rows.Next() {
		var result Berita
		rows.Scan(
			&result.Id,
			&result.Title,
			&result.Section,
			&result.ThumbImg,
			&result.Desc,
		)
		news = append(news, &result)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, news)
}

func (c *ProfileController) GetBeritaById(ctx *gin.Context) {
	id := ctx.Param("beritaId")
	parsedId, err := strconv.Atoi(id)
	if err != nil {
		c.res.AbortInvalidBerita(ctx, err, err.Error(), nil)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	type Berita struct {
		Id       int    `json:"id"`
		Title    string `json:"title"`
		Section  string `json:"section"`
		ThumbImg string `json:"thumbImg"`
		Desc     string `json:"descriptions"`
		Details  string `json:"details"`
	}

	var news Berita
	err = c.db.QueryRowContext(_context, `
	SELECT id, title, section, thumb_img, descriptions, details FROM announcements
		WHERE id = ?`, parsedId).Scan(
		&news.Id,
		&news.Title,
		&news.Section,
		&news.ThumbImg,
		&news.Desc,
		&news.Details,
	)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err == sql.ErrNoRows {
		c.res.AbortWithStatusJSON(ctx, err, "berita not found", err.Error(), http.StatusNotFound, nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, news)
}
