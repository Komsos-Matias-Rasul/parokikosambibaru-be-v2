package controller

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *Controller) CoreGetArticleById(ctx *gin.Context) {
	articleId := ctx.Param("id")
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
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
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
	if err == sql.ErrNoRows {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": row})
}

func (c *Controller) CoreGetArticleByEdition(ctx *gin.Context) {
	editionId := ctx.Param("editionId")
	parsedEditionId, err := strconv.Atoi(editionId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid edition id"})
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
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
}
