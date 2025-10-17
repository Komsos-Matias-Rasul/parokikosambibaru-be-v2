package controller

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func (c *Controller) CoreGetAllWriters(ctx *gin.Context) {
	type writer struct {
		Id         int    `json:"id"`
		WriterName string `json:"writer_name"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	rows, err := c.db.QueryContext(_context, `SELECT * FROM writers ORDER BY id`)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()

	writers := []*writer{}
	for rows.Next() {
		var w writer
		if err := rows.Scan(&w.Id, &w.WriterName); err != nil {
			c.res.AbortDatabaseError(ctx, err, nil)
			return
		}
		writers = append(writers, &w)
	}
	if err := rows.Err(); err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, writers)
}
