package util

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func GetTimeDiffAsString(from, to time.Time) string {
	totalSecs := int64(from.Sub(to).Seconds())
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60

	return fmt.Sprintf(" %02d : %02d : %02d", hours, minutes, seconds)
}

func ParseToLocalTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
}

func PtrB(b bool) *bool {
	return &b
}

func FailOnError(err error, msg string) {
	if err == nil {
		return
	}

	logrus.Fatalf("%s : %v", msg, err)
}
