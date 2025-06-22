package lib

import (
	"log"
	"os"
)

func NewLogger(msg string) {
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	logger := log.New(file, "app:", log.LstdFlags|log.Lshortfile)
	logger.Println(msg)
}
