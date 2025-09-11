package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (c *Controller) CoreSaveImage(ctx *gin.Context) {
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

	_type := ctx.Request.FormValue("uploadType")
	if _type == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing required field. (uploadType)"})
		return
	}
	if _type != "editionCover" && _type != "articleCover" && _type != "articleContent" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid upload type. (editionCover, articleCover, articleContent)"})
		return
	}

	id := ctx.Request.FormValue("articleId")
	if id == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing required field. (articleId)"})
		return
	}
	_, err = strconv.Atoi(id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid article id."})
		return
	}

	img, err := file.Open()
	if err != nil {
		fmt.Println(err.Error())
	}
	defer img.Close()

	// client, err := lib.GetCloudStorage(ctx.Request.Context())
}
