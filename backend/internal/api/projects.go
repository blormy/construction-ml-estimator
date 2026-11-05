package httpapi

import (
	"constructmaterialai/internal/auth"
	"constructmaterialai/internal/models"
	"constructmaterialai/internal/repository"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// dto моделей проектов

type equipmentItemDTO struct {
	EquipmentType string          `json:"equipment_type"`
	Parameters    json.RawMessage `json:"parameters"`
	WeightResult  *float64        `json:"weight_result,omitempty"`
}

type createProjectRequest struct {
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	TeamID         *uint              `json:"team_id,omitempty"`
	EquipmentItems []equipmentItemDTO `json:"equipment_items"`
}

type projectListItem struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ItemCount   int       `json:"item_count"`
	TotalWeight float64   `json:"total_weight"`
	TeamID      *uint     `json:"team_id,omitempty"`
	OwnerID     uint      `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type equipmentItemResponse struct {
	ID            uint            `json:"id"`
	EquipmentType string          `json:"equipment_type"`
	Parameters    json.RawMessage `json:"parameters"`
	WeightResult  *float64        `json:"weight_result,omitempty"`
}

type projectDetailResponse struct {
	ID             uint                    `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description"`
	EquipmentItems []equipmentItemResponse `json:"equipment_items"`
	TeamID         *uint                   `json:"team_id,omitempty"`
	OwnerID        uint                    `json:"owner_id"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type moveProjectRequest struct {
	TeamID *uint `json:"team_id"`
}

// маршрутизация запросов к проектам (post и get)
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createProjectHandler(w, r)
	case http.MethodGet:
		listProjectsHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// маршрутизация запросов к проекту по id
func ProjectByIDHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	path = strings.TrimSuffix(path, "/")

	// проверяем подмаршруты: /api/projects/{id}/move
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "Неверный путь")
		return
	}

	projectID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || projectID == 0 {
		writeError(w, http.StatusBadRequest, "Неверный ID проекта")
		return
	}

	// подмаршрут /move
	if len(parts) == 2 && parts[1] == "move" {
		if r.Method == http.MethodPatch {
			moveProjectHandler(w, r, uint(projectID))
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		}
		return
	}

	// основной маршрут /api/projects/{id}
	switch r.Method {
	case http.MethodGet:
		getProjectHandler(w, r, uint(projectID))
	case http.MethodPut:
		updateProjectHandler(w, r, uint(projectID))
	case http.MethodDelete:
		deleteProjectHandler(w, r, uint(projectID))
	default:
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// получение проекта по id
func getProjectHandler(w http.ResponseWriter, r *http.Request, projectID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	project, err := repository.GetProjectByID(projectID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Проект не найден")
		return
	}

	// формируем ответ со списком оборудования
	items := make([]equipmentItemResponse, 0, len(project.EquipmentItems))
	for _, item := range project.EquipmentItems {
		items = append(items, equipmentItemResponse{
			ID:            item.ID,
			EquipmentType: item.EquipmentType,
			Parameters:    json.RawMessage(item.Parameters),
			WeightResult:  item.WeightResult,
		})
	}

	writeJSON(w, http.StatusOK, projectDetailResponse{
		ID:             project.ID,
		Name:           project.Name,
		Description:    project.Description,
		EquipmentItems: items,
		TeamID:         project.TeamID,
		OwnerID:        project.UserID,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
	})
}

// удаление проекта по id
func deleteProjectHandler(w http.ResponseWriter, r *http.Request, projectID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	if err := repository.DeleteProject(projectID, userID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Проект удалён"})
}

type updateProjectRequest struct {
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	EquipmentItems []equipmentItemDTO `json:"equipment_items"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// обновление данных проекта
func updateProjectHandler(w http.ResponseWriter, r *http.Request, projectID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// загружаем проект для проверки прав доступа
	existingProj, err := repository.GetProjectByID(projectID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Проект не найден или нет доступа")
		return
	}

	// Только владелец может обновлять? Для командного проекта могут обновлять все участники.
	// GetProjectByID уже проверяет: user_id = ? ИЛИ team_id IN (команды пользователя)
	
	project := models.Project{
		ID:          projectID,
		UserID:      existingProj.UserID,
		TeamID:      existingProj.TeamID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
	}

	for _, dto := range req.EquipmentItems {
		project.EquipmentItems = append(project.EquipmentItems, models.EquipmentItem{
			ProjectID:     projectID,
			EquipmentType: dto.EquipmentType,
			Parameters:    models.JSONB(dto.Parameters),
			WeightResult:  dto.WeightResult,
		})
	}

	if err := repository.UpdateProject(&project, req.UpdatedAt); err != nil {
		if err.Error() == "CONFLICT_VERSION" {
			writeError(w, http.StatusConflict, "Конфликт версий: проект был изменён другим участником.")
			return
		}
		writeError(w, http.StatusInternalServerError, "Ошибка обновления проекта")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Проект обновлён"})
}

// перенос проекта в команду
func moveProjectHandler(w http.ResponseWriter, r *http.Request, projectID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	var req moveProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// проверяем, что пользователь является участником команды
	if req.TeamID != nil && *req.TeamID > 0 {
		if !repository.IsTeamMember(*req.TeamID, userID) {
			writeError(w, http.StatusForbidden, "Вы не являетесь участником этой команды")
			return
		}
	}

	if err := repository.MoveProjectToTeam(projectID, userID, req.TeamID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Проект перенесён"})
}

// создание нового проекта со списком оборудования
func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "Название проекта обязательно")
		return
	}

	// формируем модель проекта для сохранения
	project := models.Project{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		TeamID:      req.TeamID,
	}

	for _, dto := range req.EquipmentItems {
		if strings.TrimSpace(dto.EquipmentType) == "" {
			writeError(w, http.StatusBadRequest, "Параметр equipment_type обязателен для каждого элемента")
			return
		}
		project.EquipmentItems = append(project.EquipmentItems, models.EquipmentItem{
			EquipmentType: dto.EquipmentType,
			Parameters:    models.JSONB(dto.Parameters),
			WeightResult:  dto.WeightResult,
		})
	}

	// сохранение проекта и оборудования в одной транзакции
	if err := repository.CreateProject(&project); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка сохранения проекта")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      project.ID,
		"message": "Проект успешно сохранён",
	})
}

// получение списка проектов пользователя (личные и командные)
func listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	projects, err := repository.GetProjectsByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка загрузки проектов")
		return
	}

	result := make([]projectListItem, 0, len(projects))
	for _, p := range projects {
		var totalWeight float64
		for _, item := range p.EquipmentItems {
			if item.WeightResult != nil {
				totalWeight += *item.WeightResult
			}
		}
		result = append(result, projectListItem{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			ItemCount:   len(p.EquipmentItems),
			TotalWeight: totalWeight,
			TeamID:      p.TeamID,
			OwnerID:     p.UserID,
			CreatedAt:   p.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, result)
}
