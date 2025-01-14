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
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if len(repeat) == 0 {
		return "", fmt.Errorf("[NextDate]: repeat rule can't be empty")
	}

	nowDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("[NextDate]: wrong date: %v", err)
	}

	repeatError := fmt.Errorf("[NextDate]: wrong repeat format")

	switch repeat[0] {
	case 'd':
		if days := strings.Split(repeat, " "); len(days) != 2 {
			return "", repeatError
		} else {
			d, err := strconv.Atoi(days[1])
			if err != nil {
				return "", repeatError
			}
			if d > 400 {
				return "", repeatError
			}

			nowDate = nowDate.AddDate(0, 0, d)
			for !timeDiff(nowDate, now) {
				nowDate = nowDate.AddDate(0, 0, d)
			}
			return nowDate.Format("20060102"), nil
		}
	case 'y':
		if len(repeat) != 1 {
			return "", repeatError
		}
		nowDate = nowDate.AddDate(1, 0, 0)
		for !timeDiff(nowDate, now) {
			nowDate = nowDate.AddDate(1, 0, 0)
		}
		return nowDate.Format("20060102"), nil
	case 'w':
		if days := strings.Split(repeat, " "); len(days) != 2 {
			return "", repeatError
		} else {
			weekdayToOur := map[string]int{
				"Monday":    1,
				"Tuesday":   2,
				"Wednesday": 3,
				"Thursday":  4,
				"Friday":    5,
				"Saturday":  6,
				"Sunday":    7,
			}

			weekdays := strings.Split(days[1], ",")
			wds := make(map[int]struct{})
			for _, wd := range weekdays {
				toAdd, err := strconv.Atoi(wd)
				if err != nil || toAdd <= 0 || toAdd > 7 {
					return "", repeatError
				}
				wds[toAdd] = struct{}{}
			}

			nowDate = nowDate.AddDate(0, 0, 1)
			for {
				if timeDiff(nowDate, now) {
					if _, ok := wds[weekdayToOur[nowDate.Weekday().String()]]; ok {
						break
					}
				}
				nowDate = nowDate.AddDate(0, 0, 1)
			}

			return nowDate.Format("20060102"), nil
		}
	case 'm':
		if days := strings.Split(repeat, " "); len(days) != 2 && len(days) != 3 {
			return "", repeatError
		} else if len(days) == 2 {
			mds := strings.Split(days[1], ",")
			monthDays := make(map[int]struct{})
			last, prelast := false, false
			for _, md := range mds {
				toAdd, err := strconv.Atoi(md)
				if err != nil || toAdd == 0 || toAdd < -2 || toAdd > 31 {
					return "", repeatError
				}
				if toAdd == -1 {
					last = true
				}
				if toAdd == -2 {
					prelast = true
				}
				monthDays[toAdd] = struct{}{}
			}

			nowDate = nowDate.AddDate(0, 0, 1)
			for {
				if timeDiff(nowDate, now) {

					if _, ok := monthDays[nowDate.Day()]; ok {
						break
					}

					if last && nowDate.Day() == daysInMonth(nowDate) {
						break
					}
					if prelast && nowDate.Day() == daysInMonth(nowDate)-1 {
						break
					}
				}
				nowDate = nowDate.AddDate(0, 0, 1)
			}

			return nowDate.Format("20060102"), nil
		} else if len(days) == 3 {
			mds := strings.Split(days[1], ",")
			monthDays := make(map[int]struct{})

			last, prelast := false, false
			for _, md := range mds {
				toAdd, err := strconv.Atoi(md)
				if err != nil || toAdd == 0 || toAdd < -2 || toAdd > 31 {
					return "", repeatError
				}
				if toAdd == -1 {
					last = true
				}
				if toAdd == -2 {
					prelast = true
				}
				monthDays[toAdd] = struct{}{}
			}

			m := strings.Split(days[2], ",")
			months := make(map[int]struct{})
			for _, month := range m {
				toAdd, err := strconv.Atoi(month)
				if err != nil || toAdd <= 0 || toAdd > 12 {
					return "", repeatError
				}
				months[toAdd] = struct{}{}
			}

			nowDate = nowDate.AddDate(0, 0, 1)
			for {
				if timeDiff(nowDate, now) {

					if _, ok := months[int(nowDate.Month())]; ok {
						if _, ok = monthDays[nowDate.Day()]; ok {
							break
						}
						// случаи, когда нужен последний или предпоследний день(-1 или -2)
						if last && nowDate.Day() == daysInMonth(nowDate) {
							break
						}
						if prelast && nowDate.Day() == daysInMonth(nowDate)-1 {
							break
						}
					}
				}
				nowDate = nowDate.AddDate(0, 0, 1)
			}

			return nowDate.Format("20060102"), nil
		}
	default:
		return "", repeatError
	}

	return "", nil
}
