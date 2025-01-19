package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"final_project/auth"
	db "final_project/database"
	"final_project/handlers"
	"final_project/tests"

	"github.com/joho/godotenv"
)

// startServer() обеспечивает начало работы сервера, создание БД и настройку API
func startServer() {
	ports := ":" + strconv.Itoa(tests.Port)
	webDir := "./web"

	if p := os.Getenv("PORT"); p != "" {
		ports = ":" + p
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("ошибка загрузки .env файла: ", err)
	}

	err = db.InitializeDB()
	if err != nil {
		log.Fatal("ошибка при инициализации ДБ: ", err)
	}
	defer db.DBconn.Close()

	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	http.HandleFunc("/api/task", auth.Auth(handlers.TaskHandler))
	http.HandleFunc("/api/tasks", auth.Auth(handlers.TasksHandler))
	http.HandleFunc("/api/task/done", auth.Auth(handlers.TaskDoneHandler))
	http.HandleFunc("/api/signin", handlers.SignInHandler)

	err = http.ListenAndServe(ports, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("can't start a server: %v", err))
		return
	}
}

func main() {
	startServer()
}
