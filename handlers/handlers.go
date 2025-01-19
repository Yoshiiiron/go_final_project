package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"final_project/database"
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
	rw.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, msg)))
	rw.WriteHeader(http.StatusBadRequest)
}

func handledbError(rw http.ResponseWriter, err error) {
	respondWithError(rw, fmt.Sprintf("ошибка работы с БД %w", err))
}

func queryRows(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
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
	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

}

// TasksHandler обрабатывает GET-запросы по адресу /api/tasks
func TasksHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	toSearch := r.FormValue("search")
	db := database.DBconn
	defer db.Close()

	var (
		query string
		rows  *sql.Rows
		err   error
	)

	if toSearch == "" {
		query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit`
		rows, err = queryRows(db, query, sql.Named("limit", TaskLimit))
	} else {
		query, rows, err = buildSearchQuery(db, toSearch)
	}

	if err != nil {
		handledbError(rw, err)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			handledbError(rw, err)
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
func buildSearchQuery(db *sql.DB, toSearch string) (string, *sql.Rows, error) {
	var query string
	var rows *sql.Rows
	var err error

	searchTime, err := time.Parse("02.01.2006", toSearch)
	if err == nil {
		timeToFind := searchTime.Format("20060102")
		query = `SELECT * FROM scheduler WHERE date = :date LIMIT :limit`
		rows, err = queryRows(db, query, sql.Named("limit", TaskLimit), sql.Named("date", timeToFind))
	} else {
		query = `SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit`
		rows, err = queryRows(db, query, sql.Named("limit", TaskLimit), sql.Named("search", "%"+toSearch+"%"))
	}

	return query, rows, err
}

// respondWithJSON отправляет ответ в формате JSON
func respondWithJSON(rw http.ResponseWriter, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка сериализации: %w", err))
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

	db := database.DBconn
	defer db.Close()

	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка преобразования id: %w", err))
		return
	}

	task, err := getTaskByID(db, idInt)
	if err != nil {
		respondWithError(rw, err.Error())
		return
	}

	if len(task.Repeat) == 0 {
		deleteTaskByID(db, idInt, rw)
	} else {
		updateTaskDate(db, idInt, task, rw)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{}`))
}

func getTaskByID(db *sql.DB, id int) (Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id`
	row := db.QueryRow(query, sql.Named("id", id))

	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return task, fmt.Errorf("запись не найдена")
		}
		return task, fmt.Errorf("ошибка работы с БД: %w", err)
	}
	return task, nil
}

func deleteTaskByID(db *sql.DB, id int, rw http.ResponseWriter) {
	query := `DELETE FROM scheduler WHERE id = :id`
	res, err := db.Exec(query, sql.Named("id", id))
	if err != nil {
		handledbError(rw, err)
		return
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		respondWithError(rw, "задача не найдена")
		return
	}
}

func updateTaskDate(db *sql.DB, id int, task Task, rw http.ResponseWriter) {
	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		respondWithError(rw, fmt.Sprintf("ошибка обновления даты: %w", err))
		return
	}

	query := `UPDATE scheduler SET date = :date WHERE id = :id`
	res, err := db.Exec(query, sql.Named("id", id), sql.Named("date", nextDate))
	if err != nil {
		handledbError(rw, err)
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
		respondWithError(rw, fmt.Sprintf("ошибка десериализации: %w", err))
		return
	}
	defer r.Body.Close()

	if p.Password == os.Getenv("TODO_PASSWORD") {
		token, err := generateJWTToken()
		if err != nil {
			respondWithError(rw, fmt.Sprintf("ошибка генерации токена: %w", err))
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
