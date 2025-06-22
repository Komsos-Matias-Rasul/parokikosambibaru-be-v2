package lib

import (
	"log"
	"time"
)

func Base64ToTime(encodedTimeByte []uint8) *time.Time {
	if len(encodedTimeByte) < 1 {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", string(encodedTimeByte))
	if err != nil {
		log.Println("Failed to parse time:", err.Error())
		return nil
	}
	return &t
}
