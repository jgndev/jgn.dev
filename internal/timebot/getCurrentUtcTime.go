package timebot

import "time"

func GetCurrentUtcTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}
