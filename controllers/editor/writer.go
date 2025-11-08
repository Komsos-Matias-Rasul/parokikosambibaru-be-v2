package editor

import (
	"context"
	"net/http"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *EditorController) GetAllWriters(ctx *gin.Context) {
	type Writer struct {
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

	writers := []*Writer{}
	for rows.Next() {
		var w Writer
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

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{"writers": writers})
}

func (c *EditorController) CreateWriter(ctx *gin.Context) {
	if ctx.Request.Body == nil {
		c.res.AbortInvalidRequestBody(ctx, lib.ErrInvalidBody, "missing request body", nil)
		return
	}

	type reqBody struct {
		Writer string `json:"writer"`
	}

	var payload reqBody
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	_, err := c.db.ExecContext(_context, "INSERT INTO writers (writer_name) VALUES (?)", payload.Writer)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, payload)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	c.res.SuccessWithStatusJSON(ctx, http.StatusCreated, payload, gin.H{"message": "writer created successfully"})
}
