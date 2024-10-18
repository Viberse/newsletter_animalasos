package tools

import (
	"log"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func ParseStrDateFromDB(date string) (time.Time, error) {
	date2 := date[:10] + "T" + date[11:]
	return time.Parse(time.RFC3339, date2)
}
