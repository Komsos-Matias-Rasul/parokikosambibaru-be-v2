package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) GetArticlesByCategory(ctx *gin.Context) {
	category := ctx.Query("category")
	if category == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing required query (?category=)"})
		return
	}
	_catId, err := strconv.Atoi(category)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}
	rows, err := c.db.Query(`
		SELECT a.id, a.title, slug, w.writer_name, published_date, thumb_img, a.thumb_text, a.edition_id, e.edition_year
		FROM articles as a
		JOIN writers w ON w.id = a.writer_id
		JOIN editions e ON e.id = a.edition_id
		WHERE category_id = ? AND published_date IS NOT NULL`, _catId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type articleResponseModel struct {
		Id           int        `json:"id"`
		Title        string     `json:"title"`
		Slug         string     `json:"slug"`
		Writer       string     `json:"writer_name"`
		PublisedDate *time.Time `json:"published_date"`
		ThumbImg     string     `json:"thumb_img"`
		ThumbText    string     `json:"thumb_text"`
		EditionId    int        `json:"edition_id"`
		EditionYear  int        `json:"edition_year"`
	}

	articles := []*articleResponseModel{}

	for rows.Next() {
		var result articleResponseModel
		var publishedDate []uint8
		rows.Scan(
			&result.Id,
			&result.Title,
			&result.Slug,
			&result.Writer,
			&publishedDate,
			&result.ThumbImg,
			&result.ThumbText,
			&result.EditionId,
			&result.EditionYear,
		)
		result.PublisedDate = lib.Base64ToTime(publishedDate)
		articles = append(articles, &result)
	}

	ctx.JSON(http.StatusOK, gin.H{"data": articles})
}
