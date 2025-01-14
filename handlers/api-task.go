package handlers

import (
	"database/sql"
	"encoding/json"
	db "final_project/database"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

// taskByIdHandler() обрабатывает GET-запросы по адресу /api/task
func taskByIdHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.FormValue("id")
	if len(id) == 0 {
		rw.Write([]byte(`{"error":"не указан идентификатор"}`))
		return
	}

	dB, err := db.OpenSql()

	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}
	defer dB.Close()

	idInt, err := strconv.Atoi(id)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с id %v", err).Error())))
		return
	}
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id`
	row := dB.QueryRow(query, sql.Named("id", idInt))

	var task Task
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			rw.Write([]byte(`{"error":"запись не найдена"}`))
			return
		}
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}

	data, err := json.Marshal(task)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка сериализации %v", err).Error())))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(data)
}

// updateTaskHandler() обрабатывает PUT-запросы по адресу /api/task
func updateTaskHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	decoder := json.NewDecoder(r.Body)
	var t Task
	err := decoder.Decode(&t)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`,
			fmt.Errorf("ошибка десериализации %v", err).Error())))
		return
	}
	defer r.Body.Close()

	id, date, title, comment, repeat := t.ID, t.Date, t.Title, t.Comment, t.Repeat

	if len(title) == 0 {
		rw.Write([]byte(`{"error":"заголовок не может быть пустым"}`))
		return
	}

	if len(date) == 0 {
		date = time.Now().Format("20060102")
	}

	dateTo, err := time.Parse("20060102", date)
	if err != nil {
		rw.Write([]byte(`{"error":"некорректная дата"}`))
		return
	}

	newDate := ""
	if len(repeat) > 0 {
		newDate, err = NextDate(time.Now(), date, repeat)
		if err != nil {
			rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err.Error())))
			return
		}
		if timeDiff(time.Now(), dateTo) {
			date = newDate
		}
	}

	//если дата меньше сегодняшнего числа
	if timeDiff(time.Now(), dateTo) {
		date = time.Now().Format("20060102")
	}

	database, err := db.OpenSql()
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}

	defer database.Close()

	query := `UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id`
	res, err := database.Exec(query,
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat),
		sql.Named("id", id),
	)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}
	if rows, err := res.RowsAffected(); err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	} else if rows == 0 {
		rw.Write([]byte(`{"error":"задача не найдена"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{}`))
}

// deleteTaskHandler() обрабатывает DELETE-запросы по адресу /api/task
func deleteTaskHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.FormValue("id")
	if len(id) == 0 {
		rw.Write([]byte(`{"error":"не указан идентификатор"}`))
		return
	}

	dB, err := db.OpenSql()

	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}
	defer dB.Close()

	query := `DELETE FROM scheduler WHERE id = :id`
	res, err := dB.Exec(query, sql.Named("id", id))
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}
	if rows, err := res.RowsAffected(); err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	} else if rows == 0 {
		rw.Write([]byte(`{"error":"задача не найдена"}`))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{}`))
}

// addTaskHandler() обрабатывает POST-запросы по адресу /api/task
func addTaskHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	decoder := json.NewDecoder(r.Body)
	var t Task
	err := decoder.Decode(&t)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`,
			fmt.Errorf("ошибка десериализации %v", err).Error())))
		return
	}
	defer r.Body.Close()

	date, title, comment, repeat := t.Date, t.Title, t.Comment, t.Repeat

	if len(title) == 0 {
		rw.Write([]byte(`{"error":"заголовок не может быть пустым"}`))
		return
	}

	if len(date) == 0 {
		date = time.Now().Format("20060102")
	}

	dateTo, err := time.Parse("20060102", date)
	if err != nil {
		rw.Write([]byte(`{"error":"некорректная дата"}`))
		return
	}

	newDate := ""
	if len(repeat) > 0 {
		newDate, err = NextDate(time.Now(), date, repeat)
		if err != nil {
			rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err.Error())))
			return
		}
		if timeDiff(time.Now(), dateTo) {
			date = newDate
		}
	}

	//если дата меньше сегодняшнего числа
	if timeDiff(time.Now(), dateTo) {
		date = time.Now().Format("20060102")
	}

	database, err := db.OpenSql()
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}

	defer database.Close()

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := database.Exec(query,
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat),
	)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}
	idToAdd, err := res.LastInsertId()
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, fmt.Errorf("ошибка работы с БД %v", err).Error())))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(fmt.Sprintf(`{"id":"%d"}`, idToAdd)))
}
