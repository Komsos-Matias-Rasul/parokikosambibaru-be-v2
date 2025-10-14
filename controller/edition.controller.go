package controller

import (
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
	rows, err := c.db.Query(`
		SELECT id, title, published_at, edition_year, cover_img, thumbnail_img
		FROM editions WHERE published_at IS NOT NULL`)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
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
	c.logger.Info(ctx.Copy(), nil, editions)
}

func (c *Controller) GetEditionById(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	_id, err := strconv.Atoi(editionId)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidEdition.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	type categoryResponseModel struct {
		Id    int    `json:"id"`
		Label string `json:"label"`
		Order int    `json:"order"`
	}

	type editionResponseModel struct {
		Id          int                      `json:"id"`
		Title       string                   `json:"title"`
		EditionYear int                      `json:"edition_year"`
		CoverIng    string                   `json:"cover_img"`
		Categories  []*categoryResponseModel `json:"categories"`
	}

	var edition editionResponseModel
	err = c.db.QueryRow(`
		SELECT id, title, edition_year, cover_img
		FROM editions WHERE id = ?
	`, _id).Scan(
		&edition.Id,
		&edition.Title,
		&edition.EditionYear,
		&edition.CoverIng,
	)
	if err != nil {
		res := gin.H{"err": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	rows, err := c.db.Query(`
      SELECT c.id, c.label, c.order FROM categories c
      WHERE c.edition_id = ? ORDER BY c.order ASC`, edition.Id)
	if err != nil {
		res := gin.H{"err": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var category categoryResponseModel
		rows.Scan(&category.Id, &category.Label, &category.Order)
		edition.Categories = append(edition.Categories, &category)
	}

	ctx.JSON(200, gin.H{
		"data": edition,
	})
	c.logger.Info(ctx.Copy(), nil, edition)
}
