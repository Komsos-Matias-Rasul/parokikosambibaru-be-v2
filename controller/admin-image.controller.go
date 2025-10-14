package controller

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) CoreSaveEditionCover(ctx *gin.Context) {
	year := ctx.Query("year")
	editionId := ctx.Query("editionId")
	if year == "" || editionId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing query params: (year, editionId)"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No image sent"})
		return
	}
	if file.Size > 5<<20 { // 5MB limit
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB"})
		return
	}

	fileType := file.Header.Get("Content-Type")

	if fileType != "image/jpg" && fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. (jpg, jpeg, png, webp)"})
		return
	}

	img, err := file.Open()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer img.Close()

	ext := filepath.Ext(file.Filename)
	_id := uuid.New().String()
	fileName := fmt.Sprintf("zaitun/editions/%s/%s/cover_%s%s", year, editionId, _id, ext)

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	err = services.UploadImage(img, fileName, _context)
	if _context.Err() == context.DeadlineExceeded {
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	url := fmt.Sprintf("/api/img/%s", fileName)
	ctx.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"cover_img": url,
		},
	})
}

func (c *Controller) CoreSaveArticleCover(ctx *gin.Context) {
	year := ctx.Query("year")
	articleId := ctx.Query("articleId")
	if year == "" || articleId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing query params: (year, articleId)"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No image sent"})
		return
	}
	if file.Size > 5<<20 { // 5MB limit
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB"})
		return
	}

	fileType := file.Header.Get("Content-Type")

	if fileType != "image/jpg" && fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. (jpg, jpeg, png, webp)"})
		return
	}

	img, err := file.Open()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer img.Close()

	ext := filepath.Ext(file.Filename)
	_id := uuid.New().String()
	fileName := fmt.Sprintf("zaitun/articles/%s/%s/cover_%s%s", year, articleId, _id, ext)

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	err = services.UploadImage(img, fileName, _context)
	if _context.Err() == context.DeadlineExceeded {
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	url := fmt.Sprintf("/api/img/%s", fileName)
	ctx.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"cover_img": url,
		},
	})
}

func (c *Controller) CoreSaveArticleContent(ctx *gin.Context) {
	articleId := ctx.Query("articleId")
	if articleId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing query params: (articleId)"})
		return
	}
	parsedArticleId, err := strconv.Atoi(articleId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	_context, cancel := context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	var year *int
	err = c.db.QueryRow(`
	SELECT edition_year FROM editions
	JOIN articles ON articles.edition_id=editions.id
	WHERE articles.id=?`, parsedArticleId).Scan(&year)
	if _context.Err() == context.DeadlineExceeded {
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}

	if err == sql.ErrNoRows {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found for given id"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "No image sent"})
		return
	}
	if file.Size > 5<<20 { // 5MB limit
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB"})
		return
	}

	fileType := file.Header.Get("Content-Type")

	if fileType != "image/jpg" && fileType != "image/jpeg" && fileType != "image/png" && fileType != "image/webp" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. (jpg, jpeg, png, webp)"})
		return
	}

	img, err := file.Open()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer img.Close()

	ext := filepath.Ext(file.Filename)
	_id := uuid.New().String()
	fileName := fmt.Sprintf("zaitun/articles/%d/%s/test_%s%s", *year, articleId, _id, ext)

	_context, cancel = context.WithTimeout(ctx.Request.Context(), 10*time.Second)
	defer cancel()

	err = services.UploadImage(img, fileName, _context)
	if _context.Err() == context.DeadlineExceeded {
		ctx.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
		return
	}
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	url := fmt.Sprintf("/api/img/%s", fileName)
	ctx.JSON(http.StatusCreated, gin.H{
		"success": 1,
		"file": gin.H{
			"url": url,
		},
	})
}
