package controller

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) GetArticlesByCategory(ctx *gin.Context) {
	category := ctx.Query("category")
	if category == "" {
		c.res.AbortInvalidCategory(ctx, lib.ErrInvalidCategory, "missing required query (?category=)", nil)
		return
	}
	_catId, err := strconv.Atoi(category)
	if err != nil {
		c.res.AbortInvalidCategory(ctx, err, err.Error(), nil)
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
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
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

	c.res.SuccessWithStatusOKJSON(ctx, nil, articles)
}

func (c *Controller) GetArticleBySlug(ctx *gin.Context) {
	year := ctx.Param("year")
	editionId := ctx.Param("editionId")
	slug := ctx.Param("slug")

	parsedYear, err := strconv.Atoi(year)
	if err != nil {
		c.res.AbortInvalidYear(ctx, err, err.Error(), nil)
		return
	}
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
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
		c.res.AbortArticleNotFound(ctx, err, err.Error(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, article)
}

func (c *Controller) GetTopArticles(ctx *gin.Context) {
	editionId := ctx.Query("editionId")
	if editionId == "" {
		c.res.AbortInvalidEdition(ctx, lib.ErrInvalidEdition, "missing required query (?editionId=)", nil)
		return
	}
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	rows, err := c.db.Query(`
		SELECT a.id, a.title, slug, w.writer_name, published_date, thumb_img, a.thumb_text, a.edition_id, e.edition_year
		FROM articles as a
		JOIN writers w ON w.id = a.writer_id
		JOIN editions e ON e.id = a.edition_id
		WHERE is_top_content = true AND published_date IS NOT NULL AND a.edition_id = ?`, parsedEditionId)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
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

	c.res.SuccessWithStatusOKJSON(ctx, nil, articles)
}
