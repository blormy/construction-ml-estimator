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

// dto моделей команд

type createTeamRequest struct {
	Name string `json:"name"`
}

type teamResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uint      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type addMemberRequest struct {
	Email string `json:"email"`
}

type teamMemberResponse struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// обработка запросов к командам (post — создание, get — список)
func TeamsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createTeamHandler(w, r)
	case http.MethodGet:
		listTeamsHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
	}
}

// обработка подмаршрутов команды
func TeamsSubHandler(w http.ResponseWriter, r *http.Request) {
	// парсим путь: members, projects или удаление команды
	path := strings.TrimPrefix(r.URL.Path, "/api/teams/")
	path = strings.TrimSuffix(path, "/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "Неверный путь")
		return
	}

	teamID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Неверный ID команды")
		return
	}

	if len(parts) == 1 {
		// удаление команды
		if r.Method == http.MethodDelete {
			deleteTeamHandler(w, r, uint(teamID))
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		}
		return
	}

	subPath := parts[1]
	switch subPath {
	case "members":
		switch r.Method {
		case http.MethodPost:
			addMemberHandler(w, r, uint(teamID))
		case http.MethodGet:
			listMembersHandler(w, r, uint(teamID))
		case http.MethodDelete:
			removeMemberHandler(w, r, uint(teamID))
		default:
			writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		}
	case "projects":
		if r.Method == http.MethodGet {
			listTeamProjectsHandler(w, r, uint(teamID))
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		}
	default:
		writeError(w, http.StatusNotFound, "Маршрут не найден")
	}
}

// создание новой команды
func createTeamHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "Название команды обязательно")
		return
	}

	team := models.Team{
		Name:    name,
		OwnerID: userID,
	}

	if err := repository.CreateTeam(&team); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка создания команды")
		return
	}

	writeJSON(w, http.StatusCreated, teamResponse{
		ID:        team.ID,
		Name:      team.Name,
		OwnerID:   team.OwnerID,
		CreatedAt: team.CreatedAt,
	})
}

// получение списка команд пользователя
func listTeamsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	teams, err := repository.GetTeamsByUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка загрузки команд")
		return
	}

	result := make([]teamResponse, 0, len(teams))
	for _, t := range teams {
		result = append(result, teamResponse{
			ID:        t.ID,
			Name:      t.Name,
			OwnerID:   t.OwnerID,
			CreatedAt: t.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// удаление команды
func deleteTeamHandler(w http.ResponseWriter, r *http.Request, teamID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	if err := repository.DeleteTeam(teamID, userID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Команда удалена"})
}

// добавление участника в команду
func addMemberHandler(w http.ResponseWriter, r *http.Request, teamID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	// только владелец может добавлять участников
	if !repository.IsTeamOwner(teamID, userID) {
		writeError(w, http.StatusForbidden, "Только владелец может добавлять участников")
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "Email обязателен")
		return
	}

	// ищем пользователя по email
	user, err := repository.FindUserByEmail(email)
	if err != nil {
		writeError(w, http.StatusNotFound, "Пользователь с таким email не найден")
		return
	}

	// проверяем, не является ли пользователь уже участником
	if repository.IsTeamMember(teamID, user.ID) {
		writeError(w, http.StatusConflict, "Пользователь уже является участником команды")
		return
	}

	if err := repository.AddTeamMember(teamID, user.ID, "member"); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка добавления участника")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "Участник добавлен",
	})
}

// получение списка участников команды
func listMembersHandler(w http.ResponseWriter, r *http.Request, teamID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	// только участники команды могут видеть список
	if !repository.IsTeamMember(teamID, userID) {
		writeError(w, http.StatusForbidden, "Нет доступа к команде")
		return
	}

	members, err := repository.GetTeamMembers(teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка загрузки участников")
		return
	}

	result := make([]teamMemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, teamMemberResponse{
			UserID: m.UserID,
			Email:  m.Email,
			Role:   m.Role,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// удаление участника из команды
func removeMemberHandler(w http.ResponseWriter, r *http.Request, teamID uint) {
	currentUserID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	targetUserIDStr := r.URL.Query().Get("user_id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil || targetUserID == 0 {
		writeError(w, http.StatusBadRequest, "Неверный user_id")
		return
	}

	isOwner := repository.IsTeamOwner(teamID, currentUserID)
	isSelf := currentUserID == uint(targetUserID)

	// владелец может удалять любого, участник — только себя
	if !isOwner && !isSelf {
		writeError(w, http.StatusForbidden, "Недостаточно прав")
		return
	}

	// владелец не может покинуть команду без ее удаления
	if isOwner && isSelf {
		writeError(w, http.StatusBadRequest, "Владелец не может покинуть команду. Удалите команду.")
		return
	}

	if err := repository.RemoveTeamMember(teamID, uint(targetUserID)); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка удаления участника")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Участник удалён"})
}

// получение списка проектов команды
func listTeamProjectsHandler(w http.ResponseWriter, r *http.Request, teamID uint) {
	userID, ok := auth.GetUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Требуется авторизация")
		return
	}

	// только участники команды могут видеть проекты
	if !repository.IsTeamMember(teamID, userID) {
		writeError(w, http.StatusForbidden, "Нет доступа к команде")
		return
	}

	projects, err := repository.GetProjectsByTeam(teamID)
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
