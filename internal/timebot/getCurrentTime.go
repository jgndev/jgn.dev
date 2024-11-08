package timebot

import "time"

func GetCurrentTime(timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "Error loading timezone"
	}

	return time.Now().In(loc).Format("15:04:05")
}
