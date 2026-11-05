package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UTubeRequest struct {
	Tag              string   `json:"tag"`
	ShellDiameter    *float64 `json:"shell_diameter"`
	TubeOutDiameter  *float64 `json:"tube_out_diameter"`
	TubeLen          *float64 `json:"tube_len"`
	TubeDesPres      *float64 `json:"tube_des_pres,omitempty"`
	HeatArea         *float64 `json:"heat_area,omitempty"`
}

func UTubeHandler(w http.ResponseWriter, r *http.Request) {
	// хэндлер для получения данных о теплообменнике и расчета веса через ml-сервис
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Ошибка запроса: неподдерживаемый HTTP-метод"))
		fmt.Println("Ошибка: неподдерживаемый HTTP-метод")
		return
	}

	var reqBody UTubeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: неверный формат JSON"))
		fmt.Println("Ошибка парсинга JSON:", err.Error())
		return
	}

	if reqBody.Tag == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует обязательный параметр tag"))
		fmt.Println("Ошибка: отсутствует обязательный параметр tag")
		return
	}

	if reqBody.ShellDiameter == nil || *reqBody.ShellDiameter <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр shell_diameter"))
		fmt.Println("Ошибка: некорректный параметр shell_diameter")
		return
	}

	if reqBody.TubeOutDiameter == nil || *reqBody.TubeOutDiameter <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр tube_out_diameter"))
		fmt.Println("Ошибка: некорректный параметр tube_out_diameter")
		return
	}

	if reqBody.TubeLen == nil || *reqBody.TubeLen <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр tube_len"))
		fmt.Println("Ошибка: некорректный параметр tube_len")
		return
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка сериализации JSON:", err.Error())
		return
	}

	endpoint := "http://localhost:8000/utube/estimate"
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка создания HTTP-запроса к сервису:", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка отправки запроса в ML-сервис:", err.Error())
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка отправки ответа клиенту:", err.Error())
	}
}
