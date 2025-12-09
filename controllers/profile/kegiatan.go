package profile

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *ProfileController) GetTopKegiatan(ctx *gin.Context) {
	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := c.db.QueryContext(_context, `
	SELECT id, title, section, thumb_img, descriptions, url FROM announcements
		WHERE deleted_at is null AND publish_start <= now() AND publish_end >= now()`)
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
		Url      string `json:"url"`
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
			&result.Url,
		)
		news = append(news, &result)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, news)
}
