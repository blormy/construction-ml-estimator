package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CentrifugalCompressorRequest struct {
	Tag                      string   `json:"tag"`
	ActualGasFlowRateInlet   *float64 `json:"actual_gas_flow_rate_inlet"`
	DesignGaugePressureInlet  *float64 `json:"design_gauge_pressure_inlet"`
	DesignGaugePressureOutlet *float64 `json:"design_gauge_pressure_outlet"`
	DriverPower              *float64 `json:"driver_power,omitempty"`
}

func CentrifugalCompressorHandler(w http.ResponseWriter, r *http.Request) {
	// хэндлер для получения данных о центробежном компрессоре и расчета веса через ml-сервис
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Ошибка запроса: неподдерживаемый HTTP-метод"))
		fmt.Println("Ошибка: неподдерживаемый HTTP-метод")
		return
	}

	var reqBody CentrifugalCompressorRequest
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

	if reqBody.ActualGasFlowRateInlet == nil || *reqBody.ActualGasFlowRateInlet <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр actual_gas_flow_rate_inlet"))
		fmt.Println("Ошибка: некорректный параметр actual_gas_flow_rate_inlet")
		return
	}

	if reqBody.DesignGaugePressureInlet == nil || *reqBody.DesignGaugePressureInlet <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр design_gauge_pressure_inlet"))
		fmt.Println("Ошибка: некорректный параметр design_gauge_pressure_inlet")
		return
	}

	if reqBody.DesignGaugePressureOutlet == nil || *reqBody.DesignGaugePressureOutlet <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр design_gauge_pressure_outlet"))
		fmt.Println("Ошибка: некорректный параметр design_gauge_pressure_outlet")
		return
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка сериализации JSON:", err.Error())
		return
	}

	endpoint := "http://localhost:8000/centrifugal_compressor/estimate"
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
