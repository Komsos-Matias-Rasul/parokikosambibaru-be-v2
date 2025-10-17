package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) CoreGetArticleById(ctx *gin.Context) {
	articleId := ctx.Param("articleId")
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	type article struct {
		Title       *string `json:"title"`
		WriterId    *int    `json:"writer_id"`
		HeadlineImg *string `json:"headline_img"`
		ContentJSON *string `json:"content_json"`
		CategoryId  *int    `json:"category_id"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var row article
	err = c.db.QueryRowContext(_context, `SELECT title, writer_id, headline_img, content_json, category_id FROM articles
      WHERE articles.id = ?`, parsedArticleId).Scan(
		&row.Title,
		&row.WriterId,
		&row.HeadlineImg,
		&row.ContentJSON,
		&row.CategoryId,
	)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err == sql.ErrNoRows {
		c.res.AbortArticleNotFound(ctx, err, "", nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, row)
}

func (c *Controller) CoreGetArticleByEdition(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	type article struct {
		Id                   *string    `json:"id"`
		Title                *string    `json:"title"`
		Writer               *string    `json:"writer"`
		Category             *string    `json:"category"`
		ArticlePublishedDate *time.Time `json:"published_at"`
		EditionId            *string    `json:"edition_id"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	articles := []*article{}
	rows, err := c.db.QueryContext(_context,
		`SELECT 
        a.id, 
        a.title, 
        w.writer_name as writer,
        c.label as category, 
        a.published_date,
        e.id as edition_id
      FROM active_edition ae, articles a 
      JOIN categories c ON c.id = a.category_id 
      JOIN editions e ON e.id = a.edition_id 
      JOIN writers w ON	w.id = a.writer_id
      WHERE a.edition_id = ?`, parsedEditionId)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var article article
		var articlePublishedDate []uint8
		if err := rows.Scan(
			&article.Id,
			&article.Title,
			&article.Writer,
			&article.Category,
			&articlePublishedDate,
			&article.EditionId,
		); err != nil {
			log.Println(err.Error())
		}
		article.ArticlePublishedDate = lib.Base64ToTime(articlePublishedDate)
		articles = append(articles, &article)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, articles)
}

func (c *Controller) CoreGetDrafts(ctx *gin.Context) {
	type article struct {
		Id                 *string    `json:"id"`
		Title              *string    `json:"title"`
		Writer             *string    `json:"writer"`
		Category           *string    `json:"category"`
		ArticleUpdatedDate *time.Time `json:"updated_at"`
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	articles := []*article{}
	rows, err := c.db.QueryContext(_context,
		`SELECT articles.id, title, w.writer_name, c.label as category, updated_at FROM articles
      JOIN categories c ON c.id=articles.category_id
      JOIN writers w ON w.id=articles.writer_id
      WHERE published_date is null`)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, _context.Err(), nil)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var article article
		var articleUpdatedDate []uint8
		if err := rows.Scan(
			&article.Id,
			&article.Title,
			&article.Writer,
			&article.Category,
			&articleUpdatedDate,
		); err != nil {
			log.Println(err.Error())
		}
		article.ArticleUpdatedDate = lib.Base64ToTime(articleUpdatedDate)
		articles = append(articles, &article)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, articles)
}

func (c *Controller) CoreArchiveArticle(ctx *gin.Context) {
	articleID := ctx.Param("articleId")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	if !exists {
		c.res.AbortArticleNotFound(ctx, err, "", nil)
		return
	}

	_, err = c.db.Exec(`
		UPDATE articles
		SET archived_date = ?, published_date = NULL
		WHERE id = ?`,
		time.Now().UTC(), id,
	)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	res := gin.H{"message": "article archived successfully"}
	c.res.SuccessWithStatusJSON(ctx, http.StatusAccepted, nil, res)
}

func (c *Controller) CoreDeleteArticlePermanent(ctx *gin.Context) {
	articleID := ctx.Param("articleId")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	if !exists {
		c.res.AbortArticleNotFound(ctx, err, "", nil)
		return
	}

	_, err = c.db.Exec(`
		DELETE FROM articles
		WHERE id = ?`,
		id,
	)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	res := gin.H{"message": "article deleted successfully"}
	c.res.SuccessWithStatusJSON(ctx, http.StatusAccepted, nil, res)
}

func (c *Controller) CorePublishArticle(ctx *gin.Context) {
	articleID := ctx.Param("articleId")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	if !exists {
		c.res.AbortArticleNotFound(ctx, err, "", nil)
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
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	res := gin.H{"message": "article published successfully"}
	c.res.SuccessWithStatusJSON(ctx, http.StatusAccepted, nil, res)
}

func (c *Controller) CoreCreateArticle(ctx *gin.Context) {
	editionIdParam := ctx.Param("editionId")
	editionId, err := strconv.Atoi(editionIdParam)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	const UNCATEGORIZED = 1
	const UNKNOWN_WRITER = 1
	now := time.Now().UTC()
	article, err := c.db.Exec(`
		INSERT INTO articles (edition_id, title, category_id, writer_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, editionId, "Untitled Article", UNCATEGORIZED, UNKNOWN_WRITER, now, now)

	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	articleId64, err := article.LastInsertId()
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	articleId := int(articleId64)
	res := gin.H{"message": "article created successfully", "article_id": articleId}
	c.res.SuccessWithStatusJSON(ctx, http.StatusCreated, nil, res)
}

func (c *Controller) CoreSaveDraft(ctx *gin.Context) {
	type SaveDraftPayload struct {
		ArticleData struct {
			Content json.RawMessage `json:"content"`
		} `json:"articleData"`
		IDData int `json:"IDData"`
	}
	var payload SaveDraftPayload
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	now := time.Now().UTC()

	_, err := c.db.Exec(`
        UPDATE articles
        SET content_json = ?, updated_at = ?
        WHERE id = ?
    `, string(payload.ArticleData.Content), now, payload.IDData)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	res := gin.H{"message": "draft saved successfully", "article_id": payload.IDData}
	c.res.SuccessWithStatusOKJSON(ctx, payload, res)
}

func formatTitleToSlug(title string) string {
	slug := strings.ToLower(title)
	reg := regexp.MustCompile(`[^a-z0-9\s]+`)
	slug = reg.ReplaceAllString(slug, "")
	slug = strings.TrimSpace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")

	return slug
}

func (c *Controller) CoreSaveTWC(ctx *gin.Context) {
	type RequestPayload struct {
		Title    string `json:"title"`
		Category int    `json:"category"`
		Writer   int    `json:"writer"`
		Id       int    `json:"id"`
	}

	var payload RequestPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	slug := formatTitleToSlug(payload.Title)
	now := time.Now().UTC()

	_, err := c.db.Exec(`
		UPDATE articles
		SET updated_at = ?, title = ?, slug = ?, category_id = ?, writer_id = ?
		WHERE id = ?`,
		now,
		payload.Title,
		slug,
		payload.Category,
		payload.Writer,
		payload.Id,
	)

	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}

	res := gin.H{"message": "attributes saved successfully", "article_id": payload.Id}
	c.res.SuccessWithStatusOKJSON(ctx, payload, res)
}
