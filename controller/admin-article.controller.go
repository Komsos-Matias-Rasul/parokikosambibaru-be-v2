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
		res := gin.H{"error": lib.ErrInvalidArticle.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
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
		res := gin.H{"error": lib.ErrTimeout.Error()}
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, res)
		c.logger.Error(ctx.Copy(), _context.Err(), nil, res)
		return
	}
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

	ctx.JSON(http.StatusOK, gin.H{"data": row})
	c.logger.Info(ctx.Copy(), nil, row)
}

func (c *Controller) CoreGetArticleByEdition(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidEdition.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	type article struct {
		Id                   *string    `json:"id"`
		Title                *string    `json:"title"`
		Writer               *string    `json:"writer"`
		Category             *string    `json:"category"`
		ArticlePublishedDate *time.Time `json:"article_published_date"`
		EditionId            *string    `json:"edition_id"`
	}

	var activeEdition int
	var editionTitle string
	var editionPublishedDate []uint8

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var articles []*article
	rows, err := c.db.QueryContext(_context,
		`SELECT 
        a.id, 
        a.title, 
        w.writer_name as writer,
        c.label as category, 
        a.published_date,
        e.published_at as ed_publish_date, 
        e.title as ed_title,
        ae.edition_id as active_edition, 
        e.id as edition_id
      FROM active_edition ae, articles a 
      JOIN categories c ON c.id = a.category_id 
      JOIN editions e ON e.id = a.edition_id 
      JOIN writers w ON	w.id = a.writer_id
      WHERE a.edition_id = ?`, parsedEditionId)
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

	for rows.Next() {
		var article article
		var articlePublishedDate []uint8
		if err := rows.Scan(
			&article.Id,
			&article.Title,
			&article.Writer,
			&article.Category,
			&articlePublishedDate,
			&editionPublishedDate,
			&editionTitle,
			&activeEdition,
			&article.EditionId,
		); err != nil {
			log.Println(err.Error())
		}
		article.ArticlePublishedDate = lib.Base64ToTime(articlePublishedDate)
		articles = append(articles, &article)
	}

	type ResponseModel struct {
		Articles             []*article `json:"articles"`
		ActiveEdition        int        `json:"active_edition"`
		EditionPublishedDate *time.Time `json:"edition_published_date"`
		EditionTitle         string     `json:"edition_title"`
	}

	responseData := ResponseModel{
		Articles:      articles,
		ActiveEdition: activeEdition,
		EditionTitle:  editionTitle,
	}
	responseData.EditionPublishedDate = lib.Base64ToTime(editionPublishedDate)

	ctx.JSON(http.StatusOK, gin.H{"data": responseData})
	c.logger.Info(ctx.Copy(), nil, responseData)
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

	ctx.JSON(http.StatusOK, gin.H{"data": articles})
	c.logger.Info(ctx.Copy(), nil, articles)
}

func (c *Controller) CoreArchiveArticle(ctx *gin.Context) {
	articleID := ctx.Param("articleId")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidArticle.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if !exists {
		res := gin.H{"error": lib.ErrArticleNotFound.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), lib.ErrArticleNotFound.Error(), nil, res)
		return
	}

	_, err = c.db.Exec(`
		UPDATE articles
		SET archived_date = ?, published_date = NULL
		WHERE id = ?`,
		time.Now().UTC(), id,
	)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	res := gin.H{"message": "article archived successfully"}
	ctx.JSON(http.StatusAccepted, res)
	c.logger.Info(ctx.Copy(), nil, res)
}

func (c *Controller) CoreDeleteArticlePermanent(ctx *gin.Context) {
	articleID := ctx.Param("articleId")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidArticle.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if !exists {
		res := gin.H{"error": lib.ErrArticleNotFound.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), lib.ErrArticleNotFound.Error(), nil, res)
		return
	}

	_, err = c.db.Exec(`
		DELETE FROM articles
		WHERE id = ?`,
		id,
	)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	res := gin.H{"message": "article deleted successfully"}
	ctx.JSON(http.StatusOK, res)
	c.logger.Info(ctx.Copy(), nil, res)

}

func (c *Controller) CorePublishArticle(ctx *gin.Context) {
	articleID := ctx.Param("articleId")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidArticle.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	var exists bool
	err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}
	if !exists {
		res := gin.H{"error": lib.ErrArticleNotFound.Error()}
		ctx.AbortWithStatusJSON(http.StatusNotFound, res)
		c.logger.Error(ctx.Copy(), lib.ErrArticleNotFound.Error(), nil, res)
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
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	res := gin.H{"message": "article published successfully"}
	ctx.JSON(http.StatusOK, res)
	c.logger.Info(ctx.Copy(), nil, res)
}

func (c *Controller) CoreCreateArticle(ctx *gin.Context) {
	editionIdParam := ctx.Param("editionId")
	editionId, err := strconv.Atoi(editionIdParam)
	if err != nil {
		res := gin.H{"error": lib.ErrInvalidEdition.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
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
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	articleId64, err := article.LastInsertId()
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	articleId := int(articleId64)
	res := gin.H{"message": "article created successfully", "article_id": articleId}
	ctx.JSON(http.StatusCreated, res)
	c.logger.Info(ctx.Copy(), nil, res)
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
		res := gin.H{"error": "invalid request body", "details": err.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	now := time.Now().UTC()

	_, err := c.db.Exec(`
        UPDATE articles
        SET content_json = ?, updated_at = ?
        WHERE id = ?
    `, string(payload.ArticleData.Content), now, payload.IDData)
	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	res := gin.H{"id": payload.IDData}
	ctx.JSON(http.StatusOK, res)
	c.logger.Info(ctx.Copy(), payload, res)
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
		TitleData    string `json:"titleData"`
		CategoryData int    `json:"categoryData"`
		WriterData   int    `json:"writerData"`
		IDData       int    `json:"IDData"`
	}

	var payload RequestPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		res := gin.H{"error": "invalid request body", "details": err.Error()}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		c.logger.Error(ctx.Copy(), err, nil, res)
		return
	}

	slug := formatTitleToSlug(payload.TitleData)
	now := time.Now().UTC()

	_, err := c.db.Exec(`
		UPDATE articles
		SET updated_at = ?, title = ?, slug = ?, category_id = ?, writer_id = ?
		WHERE id = ?`,
		now,
		payload.TitleData,
		slug,
		payload.CategoryData,
		payload.WriterData,
		payload.IDData,
	)

	if err != nil {
		res := gin.H{"error": lib.ErrDatabase.Error()}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		c.logger.Error(ctx.Copy(), err, payload, res)
		return
	}

	res := gin.H{"id": payload.IDData}
	ctx.JSON(http.StatusOK, res)
	c.logger.Info(ctx.Copy(), payload, res)
}
