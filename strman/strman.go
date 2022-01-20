package strman

import (
	"strconv"
	"strings"
)

func LongLatSplitter(strLongLat string) (long float64, lat float64, err error) {
	var longStr, latStr string
	longLatStrings := strings.Split(strLongLat, ",")
	if !(len(longLatStrings) < 2) {
		longStr = longLatStrings[0]
		latStr = longLatStrings[1]
	}
	long, err = strconv.ParseFloat(longStr, 64)
	if err != nil {
		return long, lat, err
	}
	lat, err = strconv.ParseFloat(latStr, 64)
	if err != nil {
		return long, lat, err
	}
	return long, lat, nil
}
