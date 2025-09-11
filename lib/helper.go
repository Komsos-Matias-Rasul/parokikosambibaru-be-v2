package lib

import (
	"log"
	"time"
)

func Base64ToTime(encodedTimeByte []uint8) *time.Time {
	if len(encodedTimeByte) < 1 {
		return nil
	}
	var t time.Time
	t, err := time.Parse("2006-01-02 15:04:05", string(encodedTimeByte))
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", string(encodedTimeByte))
		if err != nil {
			log.Println("Failed to parse time:", err.Error())
			return nil
		}
	}
	return &t
}
