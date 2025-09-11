package controller

import (
	"log"
	"net/http"

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

	ctx.JSON(200, gin.H{
		"data": responseData,
	})
}
