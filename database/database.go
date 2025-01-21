package database

import (
	"database/sql"
	"fmt"
	"os"

	"final_project/tests"

	_ "modernc.org/sqlite"
)

var ActualDbPath string
var DBconn *sql.DB

// InitializeDB проверяет существование базы данных, создаёт её и таблицы при необходимости.
func InitializeDB() error {
	// Используем путь из переменной окружения или тестового файла
	dbPath := tests.DBFile
	if path := os.Getenv("TODO_DBFILE"); path != "" {
		dbPath = path
	}
	ActualDbPath = dbPath

	// Проверяем существование базы данных
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		// Создаём базу данных, если её нет
		if err := createAndInitializeDB(dbPath); err != nil {
			return fmt.Errorf("Ошибка при создании базы данных: %w", err)
		}
	}

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("Не удалось открыть базу данных: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return fmt.Errorf("Ошибка проверки соединения с базой данных: %w", err)
	}

	DBconn = db
	return nil
}

// createAndInitializeDB создаёт новую базу данных и инициализирует таблицы.
func createAndInitializeDB(path string) error {
	// Открываем или создаём базу данных
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("Не удалось открыть базу данных: %w", err)
	}
	defer db.Close()

	// SQL-запрос для создания таблицы и индекса
	createQuery := `
	CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "19700101",
		title VARCHAR(128) NOT NULL DEFAULT "",
		comment VARCHAR(256) NOT NULL DEFAULT "",
		repeat VARCHAR(128) NOT NULL DEFAULT ""
	);
	CREATE INDEX date_scheduler on scheduler (date);
	`

	// Выполняем запрос на создание таблицы
	_, err = db.Exec(createQuery)
	if err != nil {
		return fmt.Errorf("Не удалось создать таблицу: %w", err)
	}

	return nil
}
