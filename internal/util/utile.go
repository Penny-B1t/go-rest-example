package util

import (
	"crypto/sha256"
	"errors"
	"io"
	"os"
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

// 파일 경로를 검증하는 함수
// 파일을 읽어 복사본을 생성하여 문제가 없는지 검증 
// https://stackoverflow.com/questions/35231846/golang-check-if-string-is-valid-path
func PathValid(fp string) error {
	// Check if file already exists
	if _, err := os.Stat(fp); err != nil {
		return errors.New("invaild filePath")
	}
	
	// Attempt to create it
	var d []byte
	if err := os.WriteFile(fp, d, 0644); err != nil {
		// 커스텀 에러 선언 필요 
		return errors.New("invaild filePath")
	}
	
	os.Remove(fp) // And delete it
	return nil
}


// 추가 인증 절차를 위한 파일 hash 값 추출 
// https://stackoverflow.com/questions/15879136/how-to-calculate-sha256-file-checksum-in-go
func CheckSum(path string) (string, error ){
	f, err := os.Open("file.txt")
	if err != nil {
		// 커스텀 에러 선언 필요 
		return "", err
	}
	defer f.Close()
  
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		// 커스텀 에러 선언 필요 
	  return "", err
	}

	return string(h.Sum(nil)), nil
}