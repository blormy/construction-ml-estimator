package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ConveyorRequest struct {
	Tag            string   `json:"tag"`             //обязательное
	ConveyorLength *float64 `json:"conveyor_length"` // обязательное
	BeltWidth      *float64 `json:"belt_width"`      // обязательное

	ConveyorFlowRate     *float64 `json:"conveyor_flow_rate,omitempty"`
	DriverPowerPerSecond *float64 `json:"driver_power_per_second,omitempty"`
	ConveyorSpeed        *float64 `json:"conveyor_speed,omitempty"`
	NumberOfWalkways     *float64 `json:"number_of_walkways,omitempty"`
}

func ConveyorHandler(w http.ResponseWriter, r *http.Request) {
	// хэндлер для получения данных о конвейере и расчета веса через ml-сервис
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Ошибка запроса: неподдерживаемый HTTP-метод"))
		fmt.Println("Ошибка: неподдерживаемый HTTP-метод")
		return
	}

	var reqBody ConveyorRequest
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

	if reqBody.ConveyorLength == nil || *reqBody.ConveyorLength <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр conveyor_length"))
		fmt.Println("Ошибка: некорректный параметр conveyor_length")
		return
	}

	if reqBody.BeltWidth == nil || *reqBody.BeltWidth <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр belt_width"))
		fmt.Println("Ошибка: некорректный параметр belt_width")
		return
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка сериализации JSON:", err.Error())
		return
	}

	endpoint := "http://localhost:8000/conveyor/estimate"
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
