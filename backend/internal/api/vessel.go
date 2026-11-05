package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VesselRequest struct {
	Tag                          string   `json:"tag"`                              //обязательное
	VesselDiameter               *float64 `json:"vessel_diameter"`                  // обязательное
	VesselTangentToTangentHeight *float64 `json:"vessel_tangent_to_tangent_height"` // обязательное

	DesignGaugePressure  *float64 `json:"design_gauge_pressure,omitempty"`
	DesignTemperature    *float64 `json:"design_temperature,omitempty"`
	OperatingTemperature *float64 `json:"operating_temperature,omitempty"`
	SkirtHeight          *float64 `json:"skirt_height,omitempty"`
	VesselLegHeight      *float64 `json:"vessel_leg_height,omitempty"`
}

func VesselHandler(w http.ResponseWriter, r *http.Request) {
	// хэндлер для получения данных о сосуде и расчета веса через ml-сервис
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Ошибка запроса: неподдерживаемый HTTP-метод"))
		fmt.Println("Ошибка: неподдерживаемый HTTP-метод")
		return
	}

	var reqBody VesselRequest
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

	if reqBody.VesselDiameter == nil || *reqBody.VesselDiameter <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр vessel_diameter"))
		fmt.Println("Ошибка: некорректный параметр vessel_diameter")
		return
	}

	if reqBody.VesselTangentToTangentHeight == nil || *reqBody.VesselTangentToTangentHeight <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Ошибка запроса: отсутствует или некорректен параметр vessel_tangent_to_tangent_height"))
		fmt.Println("Ошибка: некорректный параметр vessel_tangent_to_tangent_height")
		return
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Внутренняя ошибка сервера"))
		fmt.Println("Ошибка сериализации JSON:", err.Error())
		return
	}

	endpoint := "http://localhost:8000/vessel/estimate"
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
