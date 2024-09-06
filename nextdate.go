package main

import (
	"errors"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
)

func nextDate(now time.Time, date string, repeat string) (string, error) {
	dateTime, err := time.Parse("20060102", date)
	if err != nil {
		log.Println(err)
		return formatError()
	}
	if len(repeat) == 0 {
		return "", errors.New("не заполнен repeat")
	}
	repeatArr := strings.Split(repeat, " ")
	switch repeatArr[0] {
	case "y":
		return addYear(now, dateTime), nil
	case "m":
		return addMonth(now, dateTime, repeatArr)
	case "w":
		return addWeek(now, dateTime, repeatArr)
	case "d":
		return addDay(now, dateTime, repeatArr)
	default:
		return formatError()
	}
}

func addYear(now, date time.Time) string {
	date = date.AddDate(1, 0, 0)
	for date.Before(now) || date.Equal(now) {
		date = date.AddDate(1, 0, 0)
	}
	return date.Format("20060102")
}

func addMonth(now, date time.Time, repeat []string) (string, error) {
	newDate := date
	if now.Before(date) {
		now = date
	}
	switch len(repeat) {
	case 2:
		rull := strings.Split(repeat[1], ",")
		for newDate.Before(now) || newDate.Equal(now) {
			intDays := make([]int, len(rull))
			for i, rullDay := range rull {
				rullDayInt, err := strconv.Atoi(rullDay)
				if err != nil {
					log.Println(err)
					return formatError()
				}
				if rullDayInt > 31 || rullDayInt < -31 {
					return formatError()
				}
				if rullDayInt < 0 {
					tempDate := timeDate(newDate.Year(), newDate.Month()+1, 1, newDate.Location())
					rullDayInt = tempDate.AddDate(0, 0, rullDayInt).Day()
				}
				intDays[i] = rullDayInt
			}
			slices.Sort(intDays)
			newDateSet := false
			for _, intDay := range intDays {
				tempDate := time.Date(newDate.Year(), newDate.Month(), intDay, 0, 0, 0, 0, newDate.Location())
				if tempDate.Month() != newDate.Month() { // если число перескакивает на следующий месяц
					newDate = timeDate(newDate.Year(), newDate.Month()+1, intDays[0], newDate.Location())
					newDateSet = true
					break
				}
				if tempDate.After(newDate) {
					newDate = tempDate
					break
				}
			}
			if newDate.After(now) {
				return newDate.Format("20060102"), nil
			}
			if !newDateSet {
				newDate = timeDate(newDate.Year(), newDate.Month()+1, intDays[0], newDate.Location())
			}

		}
	case 3:
		rull := strings.Split(repeat[1], ",")
		rullMonths := strings.Split(repeat[2], ",")
		for newDate.Before(now) || newDate.Equal(now) {

			intMonths := make([]int, len(rullMonths))
			for i, rullMonth := range rullMonths {
				rullMonthInt, err := strconv.Atoi(rullMonth)
				if err != nil {
					log.Println(err)
					return formatError()
				}
				if rullMonthInt > 12 {
					return formatError()
				}
				if rullMonthInt < 0 {
					return formatError()
				}
				intMonths[i] = rullMonthInt
			}
			slices.Sort(intMonths)
			firstDay := 1
			newDateSet := false
			for _, intMonth := range intMonths {

				intDays := make([]int, len(rull))
				for i, rullDay := range rull {
					rullDayInt, err := strconv.Atoi(rullDay)
					if err != nil {
						log.Println(err)
						return formatError()
					}
					if rullDayInt > 31 || rullDayInt < -31 {
						return formatError()
					}
					if rullDayInt < 0 {
						tempDate := timeDate(newDate.Year(), time.Month(intMonth)+1, 1, newDate.Location())
						rullDayInt = tempDate.AddDate(0, 0, rullDayInt).Day()
					}
					intDays[i] = rullDayInt
				}
				slices.Sort(intDays)
				firstDay = intDays[0]
				for _, intDay := range intDays {
					tempDate := timeDate(newDate.Year(), time.Month(intMonth), intDay, newDate.Location())
					if tempDate.Month() != time.Month(intMonth) { // если число перескакивает на следующий месяц
						break
					}
					if tempDate.After(newDate) {
						newDate = tempDate
						newDateSet = true
						break
					}
				}
				if newDateSet {
					break
				}

			}
			if newDate.Before(now) || newDate.Equal(now) {
				newDate = timeDate(newDate.Year()+1, time.Month(intMonths[0]), firstDay, newDate.Location())
			}

		}
	default:
		return formatError()
	}
	return newDate.Format("20060102"), nil
}

func addWeek(now, date time.Time, repeat []string) (string, error) {
	newDate := date
	if now.Before(date) {
		now = date
	}
	switch len(repeat) {
	case 2:
		rull := strings.Split(repeat[1], ",")
		intDays := make([]int, len(rull))
		for i, rullDay := range rull {
			rullDayInt, err := strconv.Atoi(rullDay)
			if err != nil {
				log.Println(err)
				return formatError()
			}
			if rullDayInt > 7 || rullDayInt < 1 {
				return formatError()
			}
			if rullDayInt == 7 {
				rullDayInt = 0
			}
			intDays[i] = rullDayInt
		}
		slices.Sort(intDays)
		tempDate := newDate
		for newDate.Before(now) || newDate.Equal(now) {
			tempDate = tempDate.AddDate(0, 0, 1)
			if slices.Contains(intDays, int(tempDate.Weekday())) {
				newDate = tempDate
			}
		}
	default:
		return formatError()
	}
	return newDate.Format("20060102"), nil
}

func addDay(now, date time.Time, repeat []string) (string, error) {
	if len(repeat) != 2 {
		return formatError()
	}
	rull := repeat[1]
	days, err := strconv.Atoi(rull)
	if err != nil {
		log.Println(err)
		return formatError()
	}
	if days > 400 {
		return formatError()
	}
	date = date.AddDate(0, 0, days)
	for date.Before(now) || date.Equal(now) {
		date = date.AddDate(0, 0, days)
	}
	return date.Format("20060102"), nil
}

func formatError() (string, error) {
	return "", errors.New("ошибка формата")
}

func timeDate(year int, month time.Month, day int, location *time.Location) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}
