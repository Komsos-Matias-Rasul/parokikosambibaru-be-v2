package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) GetArticlesByCategory(ctx *gin.Context) {
	category := ctx.Query("category")
	if category == "" {
		res := gin.H{"error": "missing required query (?category=)"}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), errors.New("missing required query (?category=)"), nil, res)
		return
	}
	_catId, err := strconv.Atoi(category)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidCategory.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := c.db.QueryContext(_context, `
	SELECT a.id, a.title, slug, w.writer_name, published_date, thumb_img, a.thumb_text, a.edition_id, e.edition_year
	FROM articles as a
	JOIN writers w ON w.id = a.writer_id
	JOIN editions e ON e.id = a.edition_id
	WHERE category_id = ? AND published_date IS NOT NULL`, _catId)
	if _context.Err() == context.DeadlineExceeded {
		res := gin.H{"error": lib.ErrTimeout.Error()}
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, res)
		c.logger.Error(ctx.Copy(), _context.Err(), nil, res)
		return
	}
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
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
	c.logger.Info(ctx.Copy(), nil, articles)
}

func (c *Controller) GetArticleBySlug(ctx *gin.Context) {
	year := ctx.Param("year")
	editionId := ctx.Param("editionId")
	slug := ctx.Param("slug")

	parsedYear, err := strconv.Atoi(year)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidYear.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidEdition.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
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
		res := gin.H{"error": lib.ErrArticleNotFound.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": article})
	c.logger.Info(ctx.Copy(), nil, article)
}

func (c *Controller) GetTopArticles(ctx *gin.Context) {
	editionId := ctx.Query("editionId")
	if editionId == "" {
		res := gin.H{"error": "missing required query (?editionId=)"}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), errors.New("missing required query (?editionId=)"), nil, res)
		return
	}
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidEdition.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	rows, err := c.db.Query(`
		SELECT a.id, a.title, slug, w.writer_name, published_date, thumb_img, a.thumb_text, a.edition_id, e.edition_year
		FROM articles as a
		JOIN writers w ON w.id = a.writer_id
		JOIN editions e ON e.id = a.edition_id
		WHERE is_top_content = true AND published_date IS NOT NULL AND a.edition_id = ?`, parsedEditionId)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
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
	c.logger.Info(ctx.Copy(), nil, articles)
}
