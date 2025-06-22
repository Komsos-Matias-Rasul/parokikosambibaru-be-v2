package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/models"
	"github.com/gin-gonic/gin"
)

type EditionResponseModel struct {
	Id          int        `json:"id"`
	Title       string     `json:"title"`
	CreatedAt   *time.Time `json:"createdAt"`
	PublishedAt *time.Time `json:"publishedAt"`
	EditionYear *int       `json:"editionYear"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	CoverImg    *string    `json:"coverImg"`
	Thumbnail   *string    `json:"thumbnail"`
}

func (c *Controller) GetAllEditions(ctx *gin.Context) {
	var editions []*EditionResponseModel
	rows, err := c.db.Query(`SELECT * FROM editions`)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var row models.Edition
		if err := rows.Scan(
			&row.Id,
			&row.Title,
			&row.CreatedAt,
			&row.PublishedAt,
			&row.EditionYear,
			&row.ArchivedAt,
			&row.CoverImg,
			&row.Thumbnail,
		); err != nil {
			log.Println(err.Error())
		}

		editions = append(editions, &EditionResponseModel{
			row.Id,
			row.Title,
			lib.Base64ToTime(row.CreatedAt),
			lib.Base64ToTime(row.PublishedAt),
			&row.EditionYear,
			lib.Base64ToTime(row.ArchivedAt),
			&row.CoverImg,
			&row.Thumbnail,
		})
	}

	ctx.JSON(200, gin.H{
		"data": editions,
	})
}
