package services

import (
	"log"
	"time"
)

func AddTime(unixTime int64, value int, unit string) int64 {
	t := time.Unix(unixTime, 0).UTC()
	switch unit {
	case "day":
		return t.AddDate(0, 0, value).Unix()
	case "week":
		return t.AddDate(0, 0, value*7).Unix()
	case "month":
		return t.AddDate(0, value, 0).Unix()
	case "year":
		return t.AddDate(value, 0, 0).Unix()
	default:
		log.Println("Error: unknown time unit:", unit)
		return 0
	}
}
