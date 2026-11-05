package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	pumpBackendURL       = "http://localhost:8080/pump/estimate"
	conveyorBackendURL   = "http://localhost:8080/conveyor/estimate"
	vesselBackendURL     = "http://localhost:8080/vessel/estimate"
	drumBackendURL       = "http://localhost:8080/drum/estimate"
	utubeBackendURL      = "http://localhost:8080/utube/estimate"
	towerBackendURL      = "http://localhost:8080/tower/estimate"
	boxFurnaceBackendURL            = "http://localhost:8080/box_furnace/estimate"
	centrifugalCompressorBackendURL = "http://localhost:8080/centrifugal_compressor/estimate"
)
// запросы к бэкенду
type PumpRequest struct {
	Tag         string   `json:"tag"`
	FlowRate    *float64 `json:"flow_rate,omitempty"`
	FluidHead   *float64 `json:"fluid_head,omitempty"`
	RPM         *float64 `json:"rpm,omitempty"`
	SpecGravity *float64 `json:"spec_gravity,omitempty"`
	PowerKW     *float64 `json:"power_kw,omitempty"`
}

type ConveyorRequest struct {
	Tag              string   `json:"tag"`
	ConveyorLength   *float64 `json:"conveyor_length"`
	BeltWidth        *float64 `json:"belt_width"`
	ConveyorFlowRate *float64 `json:"conveyor_flow_rate,omitempty"`
}

type VesselRequest struct {
	Tag                          string   `json:"tag"`
	VesselDiameter               *float64 `json:"vessel_diameter"`
	VesselTangentToTangentHeight *float64 `json:"vessel_tangent_to_tangent_height"`
	DesignGaugePressure          *float64 `json:"design_gauge_pressure,omitempty"`
	DesignTemperature            *float64 `json:"design_temperature,omitempty"`
	SkirtHeight                  *float64 `json:"skirt_height,omitempty"`
	VesselLegHeight              *float64 `json:"vessel_leg_height,omitempty"`
}

type DrumRequest struct {
	Tag                          string   `json:"tag"`
	VesselDiameter               *float64 `json:"vessel_diameter"`
	DesignTangentToTangentLength *float64 `json:"design_tangent_to_tangent_length"`
	DesignGaugePressure          *float64 `json:"design_gauge_pressure,omitempty"`
	DesignTemperature            *float64 `json:"design_temperature,omitempty"`
}

type UTubeRequest struct {
	Tag             string   `json:"tag"`
	ShellDiameter   *float64 `json:"shell_diameter"`
	TubeOutDiameter *float64 `json:"tube_out_diameter"`
	TubeLen         *float64 `json:"tube_len"`
	TubeDesPres     *float64 `json:"tube_des_pres,omitempty"`
	HeatArea        *float64 `json:"heat_area,omitempty"`
}

type TowerRequest struct {
	Tag                          string   `json:"tag"`
	VesselDiameter               *float64 `json:"vessel_diameter"`
	DesignTangentToTangentLength *float64 `json:"design_tangent_to_tangent_length,omitempty"`
	DesignGaugePressure          *float64 `json:"design_gauge_pressure,omitempty"`
	NumberOfTrays                *float64 `json:"number_of_trays"`
}

type PumpResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type ConveyorResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type VesselResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type DrumResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type UTubeResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type TowerResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type BoxFurnaceRequest struct {
	Tag                 string   `json:"tag"`
	Duty                *float64 `json:"duty"`
	StandardGasFlowRate *float64 `json:"standard_gas_flow_rate"`
	DesignGaugePressure *float64 `json:"design_gauge_pressure,omitempty"`
	DesignTemperature   *float64 `json:"design_temperature,omitempty"`
}

type BoxFurnaceResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

type CentrifugalCompressorRequest struct {
	Tag                       string   `json:"tag"`
	ActualGasFlowRateInlet    *float64 `json:"actual_gas_flow_rate_inlet"`
	DesignGaugePressureInlet  *float64 `json:"design_gauge_pressure_inlet"`
	DesignGaugePressureOutlet *float64 `json:"design_gauge_pressure_outlet"`
	DriverPower               *float64 `json:"driver_power,omitempty"`
}

type CentrifugalCompressorResponse struct {
	ModelVersion string  `json:"model_version"`
	Weight       float64 `json:"weight"`
}

func sendPumpToBackend(data PumpRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(pumpBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var pumpResp PumpResponse
	if err := json.Unmarshal(body, &pumpResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return pumpResp.Weight, nil
}

func sendConveyorToBackend(data ConveyorRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(conveyorBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var conveyorResp ConveyorResponse
	if err := json.Unmarshal(body, &conveyorResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return conveyorResp.Weight, nil
}

func sendVesselToBackend(data VesselRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(vesselBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var vesselResp VesselResponse
	if err := json.Unmarshal(body, &vesselResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return vesselResp.Weight, nil
}

func sendDrumToBackend(data DrumRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(drumBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var drumResp DrumResponse
	if err := json.Unmarshal(body, &drumResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return drumResp.Weight, nil
}

func sendUTubeToBackend(data UTubeRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(utubeBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var utubeResp UTubeResponse
	if err := json.Unmarshal(body, &utubeResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return utubeResp.Weight, nil
}

func sendTowerToBackend(data TowerRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(towerBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var towerResp TowerResponse
	if err := json.Unmarshal(body, &towerResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return towerResp.Weight, nil
}

func sendBoxFurnaceToBackend(data BoxFurnaceRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(boxFurnaceBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var boxFurnaceResp BoxFurnaceResponse
	if err := json.Unmarshal(body, &boxFurnaceResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return boxFurnaceResp.Weight, nil
}

func sendEquipmentToBackend(eq Equipment) (float64, error) {
	switch eq.Type {
	case "Насосы":
		req := PumpRequest{
			Tag:         eq.Tag,
			FlowRate:    eq.FlowRate,
			FluidHead:   eq.FluidHead,
			RPM:         eq.RPM,
			SpecGravity: eq.SpecGravity,
			PowerKW:     eq.PowerKW,
		}
		return sendPumpToBackend(req)

	case "Конвейер":
		req := ConveyorRequest{
			Tag:              eq.Tag,
			ConveyorLength:   eq.ConveyorLength,
			BeltWidth:        eq.BeltWidth,
			ConveyorFlowRate: eq.ConveyorFlowRate,
		}
		return sendConveyorToBackend(req)

	case "Вертикальный аппарат":
		req := VesselRequest{
			Tag:                          eq.Tag,
			VesselDiameter:               eq.VesselDiameter,
			VesselTangentToTangentHeight: eq.VesselTangentToTangentHeight,
			DesignGaugePressure:          eq.DesignGaugePressure,
			DesignTemperature:            eq.DesignTemperature,
			SkirtHeight:                  eq.SkirtHeight,
			VesselLegHeight:              eq.VesselLegHeight,
		}
		return sendVesselToBackend(req)

	case "Горизонтальная емкость":
		req := DrumRequest{
			Tag:                          eq.Tag,
			VesselDiameter:               eq.VesselDiameter,
			DesignTangentToTangentLength: eq.DesignTangentToTangentLength,
			DesignGaugePressure:          eq.DesignGaugePressure,
			DesignTemperature:            eq.DesignTemperature,
		}
		return sendDrumToBackend(req)

	case "Трубчатый теплообменник":
		req := UTubeRequest{
			Tag:             eq.Tag,
			ShellDiameter:   eq.ShellDiameter,
			TubeOutDiameter: eq.TubeOutDiameter,
			TubeLen:         eq.TubeLen,
			TubeDesPres:     eq.TubeDesPres,
			HeatArea:        eq.HeatArea,
		}
		return sendUTubeToBackend(req)

	case "Тарельчатая колонна":
		req := TowerRequest{
			Tag:                          eq.Tag,
			VesselDiameter:               eq.VesselDiameter,
			DesignTangentToTangentLength: eq.DesignTangentToTangentLength,
			DesignGaugePressure:          eq.DesignGaugePressure,
			NumberOfTrays:                eq.NumberOfTrays,
		}
		return sendTowerToBackend(req)

	case "Коробчатая технологическая печь":
		req := BoxFurnaceRequest{
			Tag:                 eq.Tag,
			Duty:                eq.Duty,
			StandardGasFlowRate: eq.StandardGasFlowRate,
			DesignGaugePressure: eq.DesignGaugePressure,
			DesignTemperature:   eq.DesignTemperature,
		}
		return sendBoxFurnaceToBackend(req)

	case "Центробежный компрессор":
		req := CentrifugalCompressorRequest{
			Tag:                       eq.Tag,
			ActualGasFlowRateInlet:    eq.ActualGasFlowRateInlet,
			DesignGaugePressureInlet:  eq.DesignGaugePressureInlet,
			DesignGaugePressureOutlet: eq.DesignGaugePressureOutlet,
			DriverPower:               eq.DriverPower,
		}
		return sendCentrifugalCompressorToBackend(req)

	default:
		return 0, fmt.Errorf("неизвестный тип оборудования: %s", eq.Type)
	}
}

func sendCentrifugalCompressorToBackend(data CentrifugalCompressorRequest) (float64, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("ошибка JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(centrifugalCompressorBackendURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("сервер (%d): %s", resp.StatusCode, string(body))
	}

	var compResp CentrifugalCompressorResponse
	if err := json.Unmarshal(body, &compResp); err != nil {
		return 0, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return compResp.Weight, nil
}
