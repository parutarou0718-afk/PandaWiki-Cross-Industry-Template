package utils

import "time"

func GetTimeHourOffset(hours int64) time.Time {
	return time.Now().Truncate(time.Hour).Add(time.Duration(hours) * time.Hour)
}
