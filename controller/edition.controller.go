package controller

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

type EditionResponseModel struct {
	Id           int        `json:"id"`
	Title        string     `json:"title"`
	PublishedAt  *time.Time `json:"published_at"`
	EditionYear  *int       `json:"edition_year"`
	CoverImg     *string    `json:"cover_img"`
	ThumbnailImg *string    `json:"thumbnail_img"`
}

func (c *Controller) GetAllEditions(ctx *gin.Context) {
	editions := []*EditionResponseModel{}
	rows, err := c.db.Query(`SELECT id, title, published_at, edition_year, cover_img, thumbnail_img FROM editions`)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var edition EditionResponseModel
		var publishedAt []uint8
		if err := rows.Scan(
			&edition.Id,
			&edition.Title,
			&publishedAt,
			&edition.EditionYear,
			&edition.CoverImg,
			&edition.ThumbnailImg,
		); err != nil {
			log.Println(err.Error())
		}
		edition.PublishedAt = lib.Base64ToTime(publishedAt)
		editions = append(editions, &edition)
	}

	ctx.JSON(200, gin.H{
		"data": editions,
	})
}

func (c *Controller) GetEditionById(ctx *gin.Context) {
	editionId := ctx.Param("id")
	_id, err := strconv.Atoi(editionId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var edition EditionResponseModel
	var publishedAt []uint8
	err = c.db.QueryRow(`
		SELECT id, title, published_at, edition_year, cover_img, thumbnail_img
		FROM editions WHERE id = ?`, _id).Scan(
		&edition.Id,
		&edition.Title,
		&publishedAt,
		&edition.EditionYear,
		&edition.CoverImg,
		&edition.ThumbnailImg,
	)
	if err == sql.ErrNoRows {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "edition not found"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": edition,
	})
}
