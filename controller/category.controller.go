package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) GetCategoriesByEdition(ctx *gin.Context) {
	editionIdParam := ctx.Query("edition")
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

func (c *Controller) GetCategoriesByArticle(ctx *gin.Context) {
	articleIdParam := ctx.Query("article")
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

	c.res.SuccessWithStatusOKJSON(ctx, nil, categories)
}
