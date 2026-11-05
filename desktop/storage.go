package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const dataFile = "projects.json"

func loadProjects() AppData {
	data := AppData{Projects: []Project{}}
	file, err := os.ReadFile(dataFile)
	if err != nil {
		return data
	}

	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Println("Ошибка загрузки данных:", err)
		return AppData{Projects: []Project{}}
	}
	return data
}

func saveProjects(data AppData) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}
	return os.WriteFile(dataFile, file, 0644)
}
