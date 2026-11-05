package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// константы cloud api

const (
	backendBaseURL = "http://localhost:8080"
	apiRegisterURL = backendBaseURL + "/api/auth/register"
	apiLoginURL    = backendBaseURL + "/api/auth/login"
	apiProjectsURL = backendBaseURL + "/api/projects"
	apiTeamsURL    = backendBaseURL + "/api/teams"
	authStateFile  = "auth_state.json"
)

// глобальное состояние авторизации

type AuthState struct {
	Token  string `json:"token"`
	Email  string `json:"email"`
	UserID uint   `json:"user_id"`
	mu     sync.RWMutex
}

var currentAuth = &AuthState{}

func (a *AuthState) IsLoggedIn() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Token != ""
}

func (a *AuthState) Set(token, email string, userID uint) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Token = token
	a.Email = email
	a.UserID = userID
}

func (a *AuthState) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Token = ""
	a.Email = ""
	a.UserID = 0
}

func (a *AuthState) GetToken() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Token
}

func (a *AuthState) GetEmail() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Email
}

// сохранение токена между сессиями

type savedAuthState struct {
	Token  string `json:"token"`
	Email  string `json:"email"`
	UserID uint   `json:"user_id"`
}

// сохранение состояния авторизации в файл
func saveAuthState() {
	state := savedAuthState{
		Token:  currentAuth.GetToken(),
		Email:  currentAuth.GetEmail(),
		UserID: currentAuth.UserID,
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(authStateFile, data, 0600)
}

// загрузка состояния авторизации из файла
func loadAuthState() {
	data, err := os.ReadFile(authStateFile)
	if err != nil {
		return
	}
	var state savedAuthState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	if state.Token != "" {
		currentAuth.Set(state.Token, state.Email, state.UserID)
	}
}

// очистка состояния авторизации и удаление файла
func clearAuthState() {
	currentAuth.Clear()
	_ = os.Remove(authStateFile)
}

// http-клиент авторизации

type authAPIResponse struct {
	Token  string `json:"token"`
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Error  string `json:"error"`
}

func doAuthRequest(url, email, password string) (*authAPIResponse, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result authAPIResponse
	_ = json.Unmarshal(raw, &result)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		msg := result.Error
		if msg == "" {
			msg = fmt.Sprintf("ошибка %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return &result, nil
}

// cloud api для проектов

type cloudEquipmentItem struct {
	EquipmentType string          `json:"equipment_type"`
	Parameters    json.RawMessage `json:"parameters"`
	WeightResult  *float64        `json:"weight_result,omitempty"`
}

type cloudCreateProjectRequest struct {
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	TeamID         *uint                `json:"team_id,omitempty"`
	EquipmentItems []cloudEquipmentItem `json:"equipment_items"`
	UpdatedAt      time.Time            `json:"updated_at,omitempty"`
}

type cloudProjectListItem struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ItemCount   int       `json:"item_count"`
	TotalWeight float64   `json:"total_weight"`
	TeamID      *uint     `json:"team_id,omitempty"`
	OwnerID     uint      `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type cloudProjectDetail struct {
	ID             uint                 `json:"id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	EquipmentItems []cloudEquipmentItem `json:"equipment_items"`
	TeamID         *uint                `json:"team_id,omitempty"`
	OwnerID        uint                 `json:"owner_id"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// отправка проекта на бэкенд с jwt токеном
func saveProjectToCloud(proj Project, token string) (uint, error) {
	items := make([]cloudEquipmentItem, 0, len(proj.Equipment))
	for _, eq := range proj.Equipment {
		params, err := json.Marshal(eq)
		if err != nil {
			return 0, fmt.Errorf("ошибка сериализации оборудования: %w", err)
		}
		w := eq.CalculatedWeight
		items = append(items, cloudEquipmentItem{
			EquipmentType: eq.Type,
			Parameters:    json.RawMessage(params),
			WeightResult:  &w,
		})
	}

	payload := cloudCreateProjectRequest{
		Name:           proj.Name,
		TeamID:         proj.TeamID,
		EquipmentItems: items,
		UpdatedAt:      proj.UpdatedAt, // для проверки конфликта версий
	}
	body, _ := json.Marshal(payload)

	method := http.MethodPost
	url := apiProjectsURL
	if proj.CloudID > 0 {
		method = http.MethodPut
		url = fmt.Sprintf("%s/%d", apiProjectsURL, proj.CloudID)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return 0, fmt.Errorf("CONFLICT_VERSION") // маркер конфликта версий для ui
	}

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(raw, &errResp)
		if errResp.Error != "" {
			return 0, fmt.Errorf("%s", errResp.Error)
		}
		return 0, fmt.Errorf("ошибка сервера: %d", resp.StatusCode)
	}
	
	var successResp struct {
		ID uint `json:"id"`
	}
	_ = json.Unmarshal(raw, &successResp)
	if successResp.ID > 0 {
		return successResp.ID, nil
	}
	
	return proj.CloudID, nil
}

// загрузка списка проектов из облака
func loadCloudProjectList(token string) ([]cloudProjectListItem, error) {
	req, _ := http.NewRequest(http.MethodGet, apiProjectsURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка авторизации (код %d)", resp.StatusCode)
	}

	var list []cloudProjectListItem
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}
	return list, nil
}

// загрузка конкретного проекта с оборудованием
func loadCloudProjectByID(projectID uint, token string) (*cloudProjectDetail, error) {
	url := fmt.Sprintf("%s/%d", apiProjectsURL, projectID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("проект не найден (код %d)", resp.StatusCode)
	}

	var detail cloudProjectDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("ошибка разбора: %w", err)
	}
	return &detail, nil
}

// конвертация облачного проекта в локальный формат
func cloudProjectToLocal(detail *cloudProjectDetail) Project {
	proj := Project{
		Name:      detail.Name,
		CloudID:   detail.ID,
		TeamID:    detail.TeamID,
		OwnerID:   detail.OwnerID,
		UpdatedAt: detail.UpdatedAt,
		Equipment: []Equipment{},
	}
	for _, item := range detail.EquipmentItems {
		var eq Equipment
		if err := json.Unmarshal(item.Parameters, &eq); err == nil {
			proj.Equipment = append(proj.Equipment, eq)
		}
	}
	return proj
}

// cloud api для команд

func createTeam(name, token string) (*CloudTeam, error) {
	body, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest(http.MethodPost, apiTeamsURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ошибка создания команды (код %d)", resp.StatusCode)
	}

	var team CloudTeam
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		return nil, fmt.Errorf("ошибка разбора: %w", err)
	}
	return &team, nil
}

func loadTeams(token string) ([]CloudTeam, error) {
	req, _ := http.NewRequest(http.MethodGet, apiTeamsURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка загрузки команд (код %d)", resp.StatusCode)
	}

	var list []CloudTeam
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("ошибка разбора: %w", err)
	}
	return list, nil
}

func addTeamMember(teamID uint, email, token string) error {
	url := fmt.Sprintf("%s/%d/members", apiTeamsURL, teamID)
	body, _ := json.Marshal(map[string]string{"email": email})
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		var errResp struct{ Error string `json:"error"` }
		_ = json.Unmarshal(raw, &errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("ошибка добавления (код %d)", resp.StatusCode)
	}
	return nil
}

func loadTeamMembers(teamID uint, token string) ([]CloudTeamMember, error) {
	url := fmt.Sprintf("%s/%d/members", apiTeamsURL, teamID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка загрузки участников (код %d)", resp.StatusCode)
	}

	var list []CloudTeamMember
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("ошибка разбора: %w", err)
	}
	return list, nil
}

func loadTeamProjects(teamID uint, token string) ([]cloudProjectListItem, error) {
	url := fmt.Sprintf("%s/%d/projects", apiTeamsURL, teamID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка загрузки проектов (код %d)", resp.StatusCode)
	}

	var list []cloudProjectListItem
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("ошибка разбора: %w", err)
	}
	return list, nil
}

func moveProjectToTeam(projectID uint, teamID *uint, token string) error {
	url := fmt.Sprintf("%s/%d/move", apiProjectsURL, projectID)
	body, _ := json.Marshal(map[string]*uint{"team_id": teamID})
	req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		var errResp struct{ Error string `json:"error"` }
		_ = json.Unmarshal(raw, &errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("ошибка переноса (код %d)", resp.StatusCode)
	}
	return nil
}

// экран авторизации

// отображение экрана входа и регистрации
func showLoginScreen(w fyne.Window) {
	title := widget.NewLabel("ConstructMaterialAI")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := widget.NewLabel("Войдите в аккаунт для сохранения проектов в облако")
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.Wrapping = fyne.TextWrapWord

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("email@example.com")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль (мин. 6 символов)")

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.Wrapping = fyne.TextWrapWord

	setStatus := func(msg string, isError bool) {
		statusLabel.SetText(msg)
	}

	doAuth := func(isRegister bool) {
		email := emailEntry.Text
		password := passwordEntry.Text
		if email == "" || password == "" {
			setStatus("Введите email и пароль", true)
			return
		}

		setStatus("Подключение...", false)

		url := apiLoginURL
		if isRegister {
			url = apiRegisterURL
		}

		go func() {
			result, err := doAuthRequest(url, email, password)
			fyne.Do(func() {
				if err != nil {
					setStatus("Ошибка: "+err.Error(), true)
					return
				}
				currentAuth.Set(result.Token, result.Email, result.UserID)
				saveAuthState()
				showProjectList(w)
			})
		}()
	}

	loginBtn := NewThemedHoverButton("Войти", nil, func() { doAuth(false) })

	registerBtn := widget.NewButton("Зарегистрироваться", func() { doAuth(true) })

	skipBtn := widget.NewButton("Продолжить без аккаунта (локальный режим)", func() {
		showProjectList(w)
	})

	// упаковка формы в карточку
	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Email", emailEntry),
			widget.NewFormItem("Пароль", passwordEntry),
		),
		statusLabel,
		container.NewGridWithColumns(2, loginBtn, registerBtn),
		widget.NewSeparator(),
		skipBtn,
	)

	// фон для карточки
	cardBg := canvas.NewRectangle(theme.Current().Color(ColorNameCardBackground, fyne.CurrentApp().Settings().ThemeVariant()))
	cardBg.CornerRadius = 12

	// контейнер формы с фиксированной шириной
	formCard := container.NewStack(
		cardBg,
		container.NewPadded(form),
	)

	// центрирование карточки с ограничением ширины
	centeredForm := container.NewCenter(container.NewGridWrap(fyne.NewSize(450, 350), formCard))

	content := container.NewVBox(
		layout.NewSpacer(),
		title,
		subtitle,
		widget.NewLabel(""),
		centeredForm,
		layout.NewSpacer(),
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
}

// диалог загрузки из облака

// диалог выбора и загрузки проекта из облака
func showCloudLoadDialog(w fyne.Window) {
	token := currentAuth.GetToken()
	if token == "" {
		dialog.ShowError(fmt.Errorf("не авторизован"), w)
		return
	}

	loadingDlg := dialog.NewCustom("Загрузка из облака", "Закрыть",
		widget.NewLabel("Получение списка проектов..."), w)
	loadingDlg.Show()

	go func() {
		list, err := loadCloudProjectList(token)
		fyne.Do(func() {
			loadingDlg.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("ошибка загрузки: %w", err), w)
				return
			}
			if len(list) == 0 {
				dialog.ShowInformation("Облако", "В облаке нет сохранённых проектов", w)
				return
			}

			// список для выбора
			names := make([]string, len(list))
			for i, p := range list {
				names[i] = fmt.Sprintf("%s (%s)", p.Name, p.CreatedAt.Format("02.01.2006"))
			}

			selectWidget := widget.NewSelect(names, nil)
			selectWidget.PlaceHolder = "Выберите проект..."

			dlg := dialog.NewCustomConfirm(
				"Загрузить из облака",
				"Загрузить", "Отмена",
				container.NewVBox(
					widget.NewLabel("Выберите проект для загрузки в локальное хранилище:"),
					selectWidget,
				),
				func(ok bool) {
					if !ok || selectWidget.SelectedIndex() < 0 {
						return
					}
					chosen := list[selectWidget.SelectedIndex()]

					go func() {
						detail, err := loadCloudProjectByID(chosen.ID, token)
						fyne.Do(func() {
							if err != nil {
								dialog.ShowError(fmt.Errorf("ошибка загрузки проекта: %w", err), w)
								return
							}
							localProj := cloudProjectToLocal(detail)
							appData := loadProjects()
							
							saveAndFinish := func() {
								if err := saveProjects(appData); err != nil {
									dialog.ShowError(err, w)
									return
								}
								dialog.ShowInformation("Успех",
									fmt.Sprintf("Проект «%s» загружен из облака", localProj.Name), w)
								showProjectList(w)
							}

							for i, existing := range appData.Projects {
								if existing.Name == localProj.Name {
									dialog.ShowConfirm("Облако",
										fmt.Sprintf("Проект «%s» уже существует локально.\nХотите перезаписать его версией из облака?", localProj.Name),
										func(ok bool) {
											if ok {
												appData.Projects[i] = localProj
												saveAndFinish()
											}
										}, w)
									return
								}
							}
							
							appData.Projects = append(appData.Projects, localProj)
							saveAndFinish()
						})
					}()
				}, w)
			dlg.Show()
		})
	}()
}
