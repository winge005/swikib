package helpers

import (
	"strconv"
	"time"
)

func GetCurrentDateTime() string {
	currentTime := time.Now()
	day := convertTwoDigits(strconv.Itoa(currentTime.Day()))
	month := convertTwoDigits(strconv.Itoa(int(currentTime.Month())))
	year := strconv.Itoa(currentTime.Year())

	hour := convertTwoDigits(strconv.Itoa(currentTime.Hour()))
	minute := convertTwoDigits(strconv.Itoa(currentTime.Minute()))
	sec := convertTwoDigits(strconv.Itoa(currentTime.Second()))

	return day + "-" + month + "-" + year + " " + hour + ":" + minute + ":" + sec
}

func GetCurrentDate() string {
	currentTime := time.Now()
	day := convertTwoDigits(strconv.Itoa(currentTime.Day()))
	month := convertTwoDigits(strconv.Itoa(int(currentTime.Month())))
	year := strconv.Itoa(currentTime.Year())
	return day + "-" + month + "-" + year

}

func GetCurrentTime() string {
	currentTime := time.Now()
	hour := convertTwoDigits(strconv.Itoa(currentTime.Hour()))
	minute := convertTwoDigits(strconv.Itoa(currentTime.Minute()))
	sec := convertTwoDigits(strconv.Itoa(currentTime.Second()))
	return hour + ":" + minute + ":" + sec
}

func convertTwoDigits(input string) string {
	if len(input) < 2 {
		return "0" + input
	}
	return input
}

func Last11Days() []string {
	var returnDates []string

	cd := GetCurrentDate()
	date, err := time.Parse("02-01-2006", cd)
	if err != nil {
		panic(err)
	}

	returnDates = append(returnDates, cd)

	for i := 1; i < 11; i++ {
		duration := time.Hour
		mp := -i * 24
		multipliedDuration := duration * time.Duration(mp)
		after := date.Add(multipliedDuration)
		yyyy := after.String()[:4]
		mmm := after.String()[5:7]
		dd := after.String()[8:10]

		returnDates = append(returnDates, dd+"-"+mmm+"-"+yyyy)
	}
	return returnDates
}
