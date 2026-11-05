package main

import (
	"constructmaterialai/internal/api"
	"constructmaterialai/internal/db"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	if err := db.Connect(); err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	fmt.Println("Успешное подключение к базе данных")

	if err := db.Migrate(); err != nil {
		log.Fatal("Ошибка миграции БД: ", err)
	}
	fmt.Println("Миграция базы данных выполнена")

	mux := http.NewServeMux()
	httpapi.RegisterRoutes(mux)

	fmt.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Ошибка запуска сервера:", err.Error())
	}
}
