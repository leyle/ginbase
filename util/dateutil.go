package util

import "time"

func GetFormatTime(format string) string {
	return time.Now().Format(format)
}