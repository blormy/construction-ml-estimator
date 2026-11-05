package httpapi

import (
	"fmt"
	"net/http"
)

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	// проверка доступности бэкенда
	_, err := w.Write([]byte("Ok"))
	if err != nil {
		fmt.Printf("Ошибка отправки ответа: %s\n", err.Error())
	} else {
		fmt.Println("Тестовый запрос успешно обработан")
	}
}
