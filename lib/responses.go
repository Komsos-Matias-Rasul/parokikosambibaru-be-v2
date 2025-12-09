package lib

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Responses struct {
	logger *Logger
}

func NewResponses(l *Logger) *Responses {
	return &Responses{
		l,
	}
}

func (r *Responses) AbortWithStatusJSON(ctx *gin.Context, err error, msg string,
	details string, status int, reqData any) {
	id := uuid.New().String()
	res := gin.H{
		"_id":       id,
		"timestamp": time.Now().UnixMilli(),
		"data": gin.H{
			"error":   msg,
			"details": details,
		},
	}

	ctx.AbortWithStatusJSON(status, res)
	r.logger.Error(ctx.Copy(), err, id, reqData, res)
}

func (r *Responses) SuccessWithStatusJSON(ctx *gin.Context, status int, reqData any, resData any) {
	id := uuid.New().String()
	res := gin.H{
		"_id":       id,
		"timestamp": time.Now().UnixMilli(),
		"data":      resData,
	}

	ctx.JSON(status, res)
	r.logger.Info(ctx.Copy(), id, reqData, res)
}

func (r *Responses) SuccessWithData(ctx *gin.Context, contentType string, data []byte, fileName string) {
	id := uuid.New().String()
	ctx.Data(http.StatusOK, contentType, data)
	r.logger.Info(ctx.Copy(), id, nil, fileName)
}

func (r *Responses) SuccessWithStatusOKJSON(ctx *gin.Context, reqData any, resData any) {
	r.SuccessWithStatusJSON(ctx, http.StatusOK, reqData, resData)
}

func (r *Responses) AbortDatabaseError(ctx *gin.Context, err error, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrDatabase.Error(),
		"", http.StatusInternalServerError, reqData)
}

func (r *Responses) AbortDatabaseTimeout(ctx *gin.Context, err error, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrTimeout.Error(),
		"", http.StatusInternalServerError, reqData)
}

func (r *Responses) AbortStorageError(ctx *gin.Context, err error, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrStorage.Error(),
		"", http.StatusInternalServerError, reqData)
}

func (r *Responses) AbortReadFailure(ctx *gin.Context, err error, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrReadFailure.Error(),
		"", http.StatusInternalServerError, reqData)
}

func (r *Responses) AbortNoObject(ctx *gin.Context, err error, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrNoObject.Error(),
		"", http.StatusNotFound, reqData)
}

func (r *Responses) AbortInvalidArticle(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrInvalidArticle.Error(),
		details, http.StatusBadRequest, reqData)
}

func (r *Responses) AbortInvalidEdition(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrInvalidEdition.Error(),
		details, http.StatusBadRequest, reqData)
}

func (r *Responses) AbortInvalidCategory(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrInvalidCategory.Error(),
		details, http.StatusBadRequest, reqData)
}

func (r *Responses) AbortInvalidYear(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrInvalidYear.Error(),
		details, http.StatusBadRequest, reqData)
}

func (r *Responses) AbortInvalidBerita(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrInvalidBerita.Error(),
		details, http.StatusBadRequest, reqData)
}

func (r *Responses) AbortInvalidRequestBody(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrInvalidBody.Error(),
		details, http.StatusBadRequest, reqData)
}

func (r *Responses) AbortArticleNotFound(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrArticleNotFound.Error(),
		details, http.StatusNotFound, reqData)
}

func (r *Responses) AbortEditionNotFound(ctx *gin.Context, err error,
	details string, reqData any) {
	r.AbortWithStatusJSON(ctx, err, ErrArticleNotFound.Error(),
		details, http.StatusNotFound, reqData)
}
