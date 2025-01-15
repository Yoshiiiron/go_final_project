package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	db "final_project/database"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	_ "modernc.org/sqlite"
)

// Лимит задач, которые будут возвращаться при поиске
const TaskLimit = 50

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func respondWithError(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, msg)))
}

func handleDBError(rw http.ResponseWriter, err error) {
	respondWithError(rw, fmt.Sprintf("ошибка работы с БД %v", err))
}

func queryRows(dB *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := dB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	return rows, nil
}

// NextDateHandler() обрабатывает GET-запросы по адресу /api/nextdate
func NextDateHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	date, repeat := r.FormValue("date"), r.FormValue("repeat")
	now, err := time.Parse("20060102", r.FormValue("now"))
	if err != nil {
		respondWithError(rw, err.Error())
		return
	}

	newDate, err := NextDate(now, date, repeat)
	if err != nil {
		respondWithError(rw, err.Error())
	} else {
		rw.Write([]byte(newDate))
	}
}

// TaskHandler() обрабатывает запросы по адресу /api/task
func TaskHandler(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(rw, r)
	case http.MethodGet:
		taskByIdHandler(rw, r)
	case http.MethodPut:
		updateTaskHandler(rw, r)
	case http.MethodDelete:
		deleteTaskHandler(rw, r)
	}
}

// TasksHandler обрабатывает GET-запросы по адресу /api/tasks
func TasksHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	toSearch := r.FormValue("search")
	dB, err := db.OpenSql()
	if err != nil {
		handleDBError(rw, err)
		return
	}
	defer dB.Close()

	var query string
	var rows *sql.Rows

	if toSearch == "" {
		query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit`
		rows, err = queryRows(dB, query, sql.Named("limit", TaskLimit))
	} else {
		query, rows, err = buildSearchQuery(dB, toSearch)
	}

	if err != nil {
		handleDBError(rw, err)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			handleDBError(rw, err)
			return
		}
		tasks = append(tasks, t)
	}

	// Если задач нет, вернуть пустой массив
	if tasks == nil {
		tasks = []Task{}
	}

	respondWithJSON(rw, struct {
		Tasks []Task `json:"tasks"`
	}{Tasks: tasks})
}

// buildSearchQuery строит SQL-запрос для поиска задач
func buildSearchQuery(dB *sql.DB, toSearch string) (string, *sql.Rows, error) {
	var query string
	var rows *sql.Rows
	var err error

	searchTime, err := time.Parse("02.01.2006", toSearch)
	if err == nil {
		timeToFind := searchTime.Format("20060102")
		query = `SELECT * FROM scheduler WHERE date = :date LIMIT :limit`
		rows, err = queryRows(dB, query, sql.Named("limit", TaskLimit), sql.Named("date", timeToFind))
	} else {
		query = `SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit`
		rows, err = queryRows(dB, query, sql.Named("limit", TaskLimit), sql.Named("search", "%"+toSearch+"%"))
	}

	return query, rows, err
}

// respondWithJSON отправляет ответ в формате JSON
func respondWithJSON(rw http.ResponseWriter, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка сериализации: %v", err))
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(response)
}

// TaskDoneHandler() обрабатывает POST-запросы по адресу /api/task/done
func TaskDoneHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.FormValue("id")
	if id == "" {
		respondWithError(rw, "не указан идентификатор")
		return
	}

	dB, err := db.OpenSql()
	if err != nil {
		handleDBError(rw, err)
		return
	}
	defer dB.Close()

	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка преобразования id: %v", err))
		return
	}

	task, err := getTaskByID(dB, idInt)
	if err != nil {
		respondWithError(rw, err.Error())
		return
	}

	if len(task.Repeat) == 0 {
		deleteTaskByID(dB, idInt, rw)
	} else {
		updateTaskDate(dB, idInt, task, rw)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{}`))
}

func getTaskByID(dB *sql.DB, id int) (Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id`
	row := dB.QueryRow(query, sql.Named("id", id))

	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, fmt.Errorf("запись не найдена")
		}
		return task, fmt.Errorf("ошибка работы с БД: %v", err)
	}
	return task, nil
}

func deleteTaskByID(dB *sql.DB, id int, rw http.ResponseWriter) {
	query := `DELETE FROM scheduler WHERE id = :id`
	res, err := dB.Exec(query, sql.Named("id", id))
	if err != nil {
		handleDBError(rw, err)
		return
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		respondWithError(rw, "задача не найдена")
		return
	}
}

func updateTaskDate(dB *sql.DB, id int, task Task, rw http.ResponseWriter) {
	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка обновления даты: %v", err))
		return
	}

	query := `UPDATE scheduler SET date = :date WHERE id = :id`
	res, err := dB.Exec(query, sql.Named("id", id), sql.Named("date", nextDate))
	if err != nil {
		handleDBError(rw, err)
		return
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		respondWithError(rw, "задача не найдена")
		return
	}
}

// SignInHandler() обрабатывает POST-запросы по адресу /api/signin
func SignInHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var p struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка десериализации: %v", err))
		return
	}
	defer r.Body.Close()

	if p.Password == os.Getenv("TODO_PASSWORD") {
		token, err := generateJWTToken()
		if err != nil {
			respondWithError(rw, fmt.Sprintf("ошибка генерации токена: %v", err))
			return
		}
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(fmt.Sprintf(`{"token":"%v"}`, token)))

	} else {
		respondWithError(rw, "Неверный пароль")
	}
}

func generateJWTToken() (string, error) {
	secret := []byte(os.Getenv("TODO_PASSWORD"))
	hash := sha256.Sum256([]byte(os.Getenv("TODO_PASSWORD")))

	claims := jwt.MapClaims{
		"hash": hex.EncodeToString(hash[:]),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("не удалось подписать JWT: %v", err)
	}

	return signedToken, nil
}
