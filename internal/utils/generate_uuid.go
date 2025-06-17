package utils

import "time"

func GenerateID() string {
	return time.Now().Format("20060102150405.000")
}
