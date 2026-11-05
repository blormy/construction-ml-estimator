package httpapi

import (
	"constructmaterialai/internal/auth"
	"constructmaterialai/internal/db"
	"constructmaterialai/internal/models"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token  string `json:"token"`
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}

// отправка JSON-ответа
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// отправка JSON-ошибки
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// регистрация нового пользователя
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email и пароль обязательны")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Пароль должен быть не менее 6 символов")
		return
	}

	// проверяем, не занят ли email
	var existing models.User
	err := db.DB.Where("email = ?", req.Email).First(&existing).Error
	if err == nil {
		writeError(w, http.StatusConflict, "Email уже зарегистрирован")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(w, http.StatusInternalServerError, "Ошибка базы данных")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка хэширования пароля")
		return
	}

	user := models.User{Email: req.Email, PasswordHash: hash}
	if err := db.DB.Create(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка создания пользователя")
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token, UserID: user.ID, Email: user.Email})
}

// авторизация пользователя
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// намеренно не уточняем: email не найден или неверный пароль
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, UserID: user.ID, Email: user.Email})
}
