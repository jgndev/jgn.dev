package timebot

import (
	"time"
)

func GetCurrentTimeCST() string {

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("3:04 PM CST")
}

func GetCurrentTime(timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "Error loading timezone"
	}

	return time.Now().In(loc).Format("15:04:05")
}

func GetCurrentUtcTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}
