package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Komsos-Matias-Rasul/parokikosambibaru-be-v2/conf"
	"github.com/gin-gonic/gin"
)

type Logger struct{}

func defaultLogger(file *os.File) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: conf.LOG_LEVEL,
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, file), opts))
	return logger
}

func (l *Logger) Debug(msg string, args slog.Attr) {
	if conf.LOG_LEVEL >= slog.LevelInfo {
		return
	}
	logfile := fmt.Sprintf("logs/%s_core_samara.log", time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()

	logger := defaultLogger(file)
	logger.Debug(msg, args)
}

func (l *Logger) Info(ctx *gin.Context, reqBody any, resBody any) {
	logfile := fmt.Sprintf("logs/%s_core_samara.log", time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()

	logger := defaultLogger(file)

	reqBytes, _ := json.Marshal(reqBody)
	resBytes, _ := json.Marshal(resBody)

	logger.Info(
		http.StatusText(ctx.Writer.Status()),
		"code", ctx.Writer.Status(),
		"address", ctx.ClientIP(),
		"source", ctx.Request.URL.String(), "request", reqBytes)
	logger.Debug("data_log", "response", resBytes)
}

func (l *Logger) Warning(status int, statusText string, source string, data string) {

}

func (l *Logger) Error(ctx *gin.Context, errors any, reqBody any, resBody any) {
	logfile := fmt.Sprintf("logs/%s_core_samara.log", time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()

	logger := defaultLogger(file)

	reqBytes, _ := json.Marshal(reqBody)
	resBytes, _ := json.Marshal(resBody)
	logger.Info(
		http.StatusText(ctx.Writer.Status()),
		"code", ctx.Writer.Status(),
		"address", ctx.ClientIP(),
		"source", ctx.Request.URL.String(),
		"error", errors,
		"request", reqBytes)
	logger.Debug("data_log", "response", resBytes)
}

func NewLogger() *Logger {
	_log := &Logger{}
	return _log
}
