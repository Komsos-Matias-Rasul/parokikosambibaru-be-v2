package controller

import (
	"context"
	"log"
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
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			log.Println(err.Error())
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

	ctx.JSON(200, gin.H{
		"data": responseData,
	})
}

func (c *Controller) CoreEditEditionInfo(ctx *gin.Context) {
	type RequestPayload struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
	}
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid edition id"})
		return
	}

	var req RequestPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
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
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
	if err != nil {
		log.Println(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": edition,
	})
}
