package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"regexp"
	"strings"

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

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := c.db.QueryContext(_context, `
		SELECT a.id, a.title, slug, w.writer_name, published_date, thumb_img, a.thumb_text, a.edition_id, e.edition_year
		FROM articles as a
		JOIN writers w ON w.id = a.writer_id
		JOIN editions e ON e.id = a.edition_id
		WHERE category_id = ? AND published_date IS NOT NULL`, _catId)
	if _context.Err() == context.DeadlineExceeded {
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
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

func (c *Controller) GetTopArticles(ctx *gin.Context) {
	editionId := ctx.Query("editionId")
	if editionId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing required query (?editionId=)"})
		return
	}
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid edition id"})
		return
	}

	rows, err := c.db.Query(`
		SELECT a.id, a.title, slug, w.writer_name, published_date, thumb_img, a.thumb_text, a.edition_id, e.edition_year
		FROM articles as a
		JOIN writers w ON w.id = a.writer_id
		JOIN editions e ON e.id = a.edition_id
		WHERE is_top_content = true AND published_date IS NOT NULL AND a.edition_id = ?`, parsedEditionId)
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

func (c *Controller) ArchiveArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	_, err = c.db.Exec(`
		UPDATE articles
		SET archived_date = ?, published_date = NULL
		WHERE id = ?`,
		time.Now().UTC(), id,
	)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to archive article"})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"message": "article archived successfully"})
}

func (c *Controller) DeleteArticlePermanent(ctx *gin.Context) {
	articleID := ctx.Param("id")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	_, err = c.db.Exec(`
		DELETE FROM articles
		WHERE id = ?`,
		id,
	)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"message": "article deleted successfully"})

}

func (c *Controller) PublishArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	now := time.Now().UTC()
	_, err = c.db.Exec(`
		UPDATE articles
		SET published_date = ?, archived_date = null
		WHERE id = ?
		`, now, id,
	)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "article published successfully"})
}

func (c *Controller) CreateArticle(ctx *gin.Context) {
	editionIdParam := ctx.Param("editionId")
	editionId, err := strconv.Atoi(editionIdParam)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid edition id"})
		return
	}

	const UNCATEGORIZED = 1
	const UNKNOWN_WRITER = 1
	now := time.Now().UTC()
	res, err := c.db.Exec(`
		INSERT INTO articles (edition_id, title, category_id, writer_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, editionId, "Untitled Article", UNCATEGORIZED, UNKNOWN_WRITER, now, now)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create article",
			"details": err.Error()})
		return
	}

	articleId64, err := res.LastInsertId()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to fetch inserted id",
			"details": err.Error()})
		return
	}

	articleId := int(articleId64)
	ctx.JSON(http.StatusCreated, gin.H{"message": "article created successfully", "article_id": articleId})
}

func (c *Controller) SaveDraft(ctx *gin.Context) {
	type SaveDraftPayload struct {
		ArticleData struct {
			Content json.RawMessage `json:"content"`
		} `json:"articleData"`
		IDData string `json:"IDData"`
	}
	var payload SaveDraftPayload
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	id, err := strconv.Atoi(payload.IDData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	now := time.Now().UTC()

	_, err = c.db.Exec(`
        UPDATE articles
        SET content_json = ?, updated_at = ?
        WHERE id = ?
    `, string(payload.ArticleData.Content), now, id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save draft", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": payload.IDData})
}

func formatTitleToSlug(title string) string {
	slug := strings.ToLower(title)
	reg := regexp.MustCompile(`[^a-z0-9\s]+`)
	slug = reg.ReplaceAllString(slug, "")
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")

	return slug
}

func (c *Controller) SaveTWC(ctx *gin.Context) {
	type RequestPayload struct {
		TitleData    string `json:"titleData"`
		CategoryData int    `json:"categoryData"`
		WriterData   int    `json:"writerData"`
		IDData       int    `json:"IDData"`
	}

	var req RequestPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	slug := formatTitleToSlug(req.TitleData)
	now := time.Now().UTC()

	_, err := c.db.Exec(`
		UPDATE articles
		SET updated_at = ?, title = ?, slug = ?, category_id = ?, writer_id = ?
		WHERE id = ?`,
		now,
		req.TitleData,
		slug,
		req.CategoryData,
		req.WriterData,
		req.IDData,
	)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update article",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": req.IDData})
}

func (c *Controller) GetCategoriesByEdition(ctx *gin.Context) {
	editionIdParam := ctx.Query("edition")
	editionId, err := strconv.Atoi(editionIdParam)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid edition id"})
		return
	}

	rows, err := c.db.Query(`
		SELECT DISTINCT c.id, c.label
		FROM categories c
		JOIN articles a ON a.category_id = c.id
		WHERE a.edition_id = ?
		ORDER BY c.label ASC
	`, editionId)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Category struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
	}

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Label); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		categories = append(categories, cat)
	}

	ctx.JSON(http.StatusOK, gin.H{"data": categories})
}
