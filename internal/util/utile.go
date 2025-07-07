package util

import (
	"strings"
	"time"
)

// FormatTimeToISO returns the time in RFC3339 format.
func FormatTimeToISO(timeToFormat time.Time) string {
	return timeToFormat.Format(time.RFC3339)
}

// FormatTimeToISO returns the time in RFC3339 format.
func CurrentISOTime() string {
	// 서울 시간대 설정
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return FormatTimeToISO(time.Now())
	}
	return FormatTimeToISO(time.Now().In(loc))
}

// IsDevMode - Checks if the given string denotes any of the development environment.
func IsDevMode(s string) bool {
	return strings.Contains(s, "local") || strings.Contains(s, "dev")

}