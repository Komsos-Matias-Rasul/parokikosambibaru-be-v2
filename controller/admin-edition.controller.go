package controller

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) CoreGetAllEditions(ctx *gin.Context) {
	editions := []*EditionResponseModel{}
	rows, err := c.db.Query(`SELECT editions.id, title, thumbnail_img, cover_img, published_at, edition_year, edition_id as active_edition FROM editions, active_edition ORDER BY created_at DESC`)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()
	var activeEdition int

	for rows.Next() {
		var edition EditionResponseModel
		var publishedAt []uint8
		if err := rows.Scan(
			&edition.Id,
			&edition.Title,
			&edition.ThumbnailImg,
			&edition.CoverImg,
			&publishedAt,
			&edition.EditionYear,
			&activeEdition,
		); err != nil {
			c.res.AbortDatabaseError(ctx, err, nil)
		}
		edition.PublishedAt = lib.Base64ToTime(publishedAt)
		editions = append(editions, &edition)
	}

	type ResponseModel struct {
		Editions      []*EditionResponseModel `json:"editions"`
		ActiveEdition int                     `json:"active_edition"`
	}

	responseData := ResponseModel{
		Editions:      editions,
		ActiveEdition: activeEdition,
	}

	time.Sleep(5 * time.Second)

	c.res.SuccessWithStatusOKJSON(ctx, nil, responseData)
}

func (c *Controller) CoreEditEditionInfo(ctx *gin.Context) {
	type RequestPayload struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
	}
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	var req RequestPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	type ResponsePayload struct {
		Title     *string `json:"title"`
		Year      *int    `json:"year"`
		EditionId *int    `json:"id"`
	}
	var edition ResponsePayload

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	_, err = c.db.QueryContext(_context,
		`
		UPDATE editions
		SET title = ?,
		edition_year = ?
		WHERE id = ?
		`, req.Title, req.Year, parsedEditionId)

	c.db.QueryRowContext(_context,
		"SELECT id, title, edition_year FROM editions WHERE id = ?",
		parsedEditionId).Scan(&edition.EditionId, &edition.Title, &edition.Year)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, req)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, req)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, req, edition)
}

func (c *Controller) CoreGetEditionInfo(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	type Edition struct {
		Id            *int       `json:"id"`
		PublishedAt   *time.Time `json:"published_at"`
		Title         *string    `json:"title"`
		ActiveEdition *int       `json:"active_edition"`
	}

	var edition Edition
	var publishedAt []uint8

	err = c.db.QueryRowContext(_context,
		`SELECT 
				e.id,
        published_at, 
        title,
        ae.edition_id as active_edition
      FROM active_edition ae, editions e
      WHERE e.id = ?`, parsedEditionId).Scan(
		&edition.Id,
		&publishedAt,
		&edition.Title,
		&edition.ActiveEdition,
	)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err == sql.ErrNoRows {
		c.res.AbortEditionNotFound(ctx, err, err.Error(), nil)
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	edition.PublishedAt = lib.Base64ToTime(publishedAt)

	c.res.SuccessWithStatusOKJSON(ctx, nil, edition)
}

func (c *Controller) CoreCreateEdition(ctx *gin.Context) {

	type ReqBody struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
	}

	var payload ReqBody
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	if payload.Year < 1970 {
		c.res.AbortInvalidRequestBody(ctx, lib.ErrInvalidBody, "year must be greater than 1970", payload)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	res, err := c.db.ExecContext(_context,
		"INSERT INTO editions (title, edition_year) VALUES (?, ?)", &payload.Title, &payload.Year)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, payload)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	time.Sleep(time.Second * 5)

	c.res.SuccessWithStatusJSON(
		ctx,
		http.StatusCreated,
		payload,
		gin.H{"message": "edition created successfully", "id": id},
	)

}
