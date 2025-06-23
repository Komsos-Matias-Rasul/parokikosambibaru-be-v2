package controller

import (
	"database/sql"
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

func (c *Controller) GetArticleBySlug(ctx *gin.Context) {
	year := ctx.Param("year")
	editionId := ctx.Param("editionId")
	slug := ctx.Param("slug")

	parsedYear, err := strconv.Atoi(year)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid edition id"})
		return
	}

	type articleResponseModel struct {
		Id           int        `json:"id"`
		Title        string     `json:"title"`
		Slug         string     `json:"slug"`
		Writer       string     `json:"writer_name"`
		PublisedDate *time.Time `json:"published_date"`
		HeadlineImg  string     `json:"headline_img"`
		Label        string     `json:"label"`
		ContentJSON  string     `json:"content_json"`
		AdsJSON      string     `json:"ads_json"`
	}

	var article articleResponseModel
	var publishedDate []uint8

	err = c.db.QueryRow(
		`
		SELECT a.id, a.title, slug, w.writer_name, published_date, headline_img, c.label,
		content_json, ads_json
		FROM articles a
		JOIN writers w ON w.id = a.writer_id
		JOIN categories c ON c.id = a.category_id
		JOIN editions e ON e.id = a.edition_id
		WHERE slug=? AND e.edition_year=? AND a.edition_id=?`,
		slug, parsedYear, parsedEditionId).Scan(
		&article.Id,
		&article.Title,
		&article.Slug,
		&article.Writer,
		&publishedDate,
		&article.HeadlineImg,
		&article.Label,
		&article.ContentJSON,
		&article.AdsJSON,
	)
	article.PublisedDate = lib.Base64ToTime(publishedDate)

	if err == sql.ErrNoRows {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": article})
}
