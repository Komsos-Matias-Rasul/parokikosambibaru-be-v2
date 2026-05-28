package editor

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/lib"
	"github.com/gin-gonic/gin"
)

func (c *EditorController) GetCategoriesByEdition(ctx *gin.Context) {
	editionIdParam := ctx.Param("editionId")
	editionId, err := strconv.Atoi(editionIdParam)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
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
		c.res.AbortDatabaseError(ctx, err, nil)
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
			c.res.AbortDatabaseError(ctx, err, nil)
			return
		}
		categories = append(categories, cat)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, categories)
}

func (c *EditorController) GetNonNullCategoriesByEdition(ctx *gin.Context) {
	editionIdParam := ctx.Param("editionId")
	editionId, err := strconv.Atoi(editionIdParam)
	if err != nil {
		c.res.AbortInvalidEdition(ctx, err, err.Error(), nil)
		return
	}

	rows, err := c.db.Query(`
		SELECT DISTINCT c.id, c.label, c.order
		FROM categories c
		JOIN editions e ON e.id = c.edition_id
		WHERE e.id = ? AND c.order IS NOT NULL
		ORDER BY c.label ASC
	`, editionId)

	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()

	type Category struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
		Order int    `json:"order"`
	}

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Label, &cat.Order); err != nil {
			c.res.AbortDatabaseError(ctx, err, nil)
			return
		}
		categories = append(categories, cat)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{"categories": categories})
}

func (c *EditorController) GetCategoriesByArticle(ctx *gin.Context) {
	articleIdParam := ctx.Param("articleId")
	articleId, err := strconv.Atoi(articleIdParam)
	if err != nil {
		c.res.AbortInvalidArticle(ctx, err, err.Error(), nil)
		return
	}

	rows, err := c.db.Query(
		"SELECT id, label, `key` FROM categories WHERE edition_id = (SELECT edition_id FROM articles WHERE articles.id = ?) OR categories.id = 1 ORDER BY id",
		articleId)
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, nil)
		return
	}
	defer rows.Close()

	type Category struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
		Key   string `json:"key"`
	}

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Label, &cat.Key); err != nil {
			c.res.AbortDatabaseError(ctx, err, nil)
			return
		}
		categories = append(categories, cat)
	}

	c.res.SuccessWithStatusOKJSON(ctx, nil, gin.H{"categories": categories})
}

func (c *EditorController) CreateCategory(ctx *gin.Context) {
	if ctx.Request.Body == nil {
		c.res.AbortInvalidRequestBody(ctx, lib.ErrInvalidBody, "missing request body", nil)
		return
	}

	type reqBody struct {
		Category  string `json:"category"`
		EditionId int    `json:"editionId"`
	}

	var payload reqBody
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	if len(strings.TrimSpace(payload.Category)) == 0 {
		err := errors.New("missing category name")
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), payload)
		return
	}

	if payload.EditionId == 0 {
		c.res.AbortInvalidEdition(ctx, lib.ErrEditionNotFound, "invalid edition", payload)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	// construct key
	ck := strings.ReplaceAll(strings.ToLower(payload.Category), " ", "_")

	_, err := c.db.ExecContext(_context, "INSERT INTO categories (label, `key`, edition_id, `order`) VALUES (?, ?, ?, ?)",
		payload.Category, ck, payload.EditionId, 0)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, payload)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	c.res.SuccessWithStatusJSON(ctx, http.StatusCreated, payload, gin.H{"message": "category created successfully"})
}

func (c *EditorController) UpdateCategoryOrder(ctx *gin.Context) {
	if ctx.Request.Body == nil {
		c.res.AbortInvalidRequestBody(ctx, lib.ErrInvalidBody, "missing request body", nil)
		return
	}

	type reqBody struct {
		Id       int `json:"id"`
		NewOrder int `json:"newOrder"`
	}

	var payload reqBody
	if err := ctx.BindJSON(&payload); err != nil {
		c.res.AbortInvalidRequestBody(ctx, err, err.Error(), nil)
		return
	}

	if payload.Id == 0 {
		err := errors.New("Invalid category")
		c.res.AbortInvalidCategory(ctx, err, err.Error(), payload)
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
	defer cancel()

	_, err := c.db.ExecContext(_context, "UPDATE categories SET `order` = ? WHERE id = ?",
		payload.NewOrder, payload.Id)
	if _context.Err() == context.DeadlineExceeded {
		c.res.AbortDatabaseTimeout(ctx, err, payload)
		return
	}
	if err != nil {
		c.res.AbortDatabaseError(ctx, err, payload)
		return
	}

	c.res.SuccessWithStatusJSON(ctx, http.StatusCreated, payload, gin.H{"message": "category order updated"})
}
