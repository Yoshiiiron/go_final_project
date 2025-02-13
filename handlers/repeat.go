package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// daysInMonth возвращает количество дней в месяце для заданной даты
func daysInMonth(t time.Time) int {
	// Возвращаем количество дней в месяце, используя стандартную библиотеку
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

// timeDiff проверяет, что дата first больше, чем дата sec
func timeDiff(first, sec time.Time) bool {
	// Сравниваем года, месяцы и дни, чтобы определить, какая дата раньше
	return first.Year() > sec.Year() ||
		(first.Year() == sec.Year() && first.Month() > sec.Month()) ||
		(first.Year() == sec.Year() && first.Month() == sec.Month() && first.Day() > sec.Day())
}

// parseRepeat разбивает строку repeat на правило и интервал, проверяя их корректность
func parseRepeat(repeat string) (string, int, error) {
	// Разбиваем строку на части по пробелам (правило и интервал)
	parts := strings.Split(repeat, " ")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("[parseRepeat]: неверный формат repeat")
	}

	// Преобразуем интервал в целое число
	interval, err := strconv.Atoi(parts[1])
	if err != nil || interval <= 0 || interval > 400 {
		return "", 0, fmt.Errorf("[parseRepeat]: неверное значение интервала")
	}
	return parts[0], interval, nil
}

// NextDate вычисляет следующую дату, соответствующую правилу repeat, начиная с текущей даты
func NextDate(now time.Time, date, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("[NextDate]: repeat rule can't be empty")
	}

	nowDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("[NextDate]: wrong date: %w", err)
	}

	repeatError := fmt.Errorf("[NextDate]: wrong repeat format")

	switch repeat[0] {
	case 'd':
		return handleDailyRepeat(nowDate, now, repeat, repeatError)
	case 'y':
		return handleYearlyRepeat(nowDate, now)
	case 'w':
		return handleWeeklyRepeat(nowDate, now, repeat, repeatError)
	case 'm':
		return handleMonthlyRepeat(nowDate, now, repeat, repeatError)
	default:
		return "", repeatError
	}
}

func handleDailyRepeat(nowDate, now time.Time, repeat string, repeatError error) (string, error) {
	days := strings.Split(repeat, " ")
	if len(days) != 2 {
		return "", repeatError
	}

	d, err := strconv.Atoi(days[1])
	if err != nil || d > 400 {
		return "", repeatError
	}

	nowDate = nowDate.AddDate(0, 0, d)
	for !timeDiff(nowDate, now) {
		nowDate = nowDate.AddDate(0, 0, d)
	}
	return nowDate.Format("20060102"), nil
}

func handleYearlyRepeat(nowDate, now time.Time) (string, error) {
	nowDate = nowDate.AddDate(1, 0, 0)
	for !timeDiff(nowDate, now) {
		nowDate = nowDate.AddDate(1, 0, 0)
	}
	return nowDate.Format("20060102"), nil
}

func handleWeeklyRepeat(nowDate, now time.Time, repeat string, repeatError error) (string, error) {
	days := strings.Split(repeat, " ")
	if len(days) != 2 {
		return "", repeatError
	}

	weekdays := parseWeekdays(days[1], repeatError)
	if weekdays == nil {
		return "", repeatError
	}

	nowDate = nowDate.AddDate(0, 0, 1)
	for {
		if timeDiff(nowDate, now) && weekdays[int(nowDate.Weekday())] {
			break
		}
		nowDate = nowDate.AddDate(0, 0, 1)
	}
	return nowDate.Format("20060102"), nil
}

func handleMonthlyRepeat(nowDate, now time.Time, repeat string, repeatError error) (string, error) {
	days := strings.Split(repeat, " ")
	if len(days) < 2 || len(days) > 3 {
		return "", repeatError
	}

	monthDays, last, prelast := parseMonthDays(days[1], repeatError)
	if monthDays == nil {
		return "", repeatError
	}

	months := parseMonths(days, repeatError)
	if len(days) == 3 && months == nil {
		return "", repeatError
	}

	nowDate = nowDate.AddDate(0, 0, 1)
	for {
		if timeDiff(nowDate, now) {
			if (len(months) == 0 || months[int(nowDate.Month())]) &&
				(monthDays[nowDate.Day()] || (last && nowDate.Day() == daysInMonth(nowDate)) ||
					(prelast && nowDate.Day() == daysInMonth(nowDate)-1)) {
				break
			}
		}
		nowDate = nowDate.AddDate(0, 0, 1)
	}
	return nowDate.Format("20060102"), nil
}

func parseWeekdays(input string, repeatError error) map[int]bool {
	weekdays := make(map[int]bool)
	for _, day := range strings.Split(input, ",") {
		wd, err := strconv.Atoi(day)
		if err != nil || wd <= 0 || wd > 7 {
			return nil
		}
		weekdays[wd] = true
	}
	return weekdays
}

func parseMonthDays(input string, repeatError error) (map[int]bool, bool, bool) {
	monthDays := make(map[int]bool)
	last, prelast := false, false

	for _, day := range strings.Split(input, ",") {
		md, err := strconv.Atoi(day)
		if err != nil || md < -2 || md > 31 {
			return nil, false, false
		}
		if md == -1 {
			last = true
		} else if md == -2 {
			prelast = true
		} else {
			monthDays[md] = true
		}
	}
	return monthDays, last, prelast
}

func parseMonths(days []string, repeatError error) map[int]bool {
	if len(days) != 3 {
		return nil
	}

	months := make(map[int]bool)
	for _, month := range strings.Split(days[2], ",") {
		m, err := strconv.Atoi(month)
		if err != nil || m < 1 || m > 12 {
			return nil
		}
		months[m] = true
	}
	return months
}
