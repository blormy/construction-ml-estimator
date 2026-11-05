package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func showStartScreen(w fyne.Window) {
	// основной заголовок
	title := canvas.NewText("ConstructMaterialAI", theme.PrimaryColor())
	title.TextSize = 52
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("Интеллектуальная система оценки веса промышленного оборудования")
	subtitle.Alignment = fyne.TextAlignCenter
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	// кнопка начала
	startBtn := NewThemedHoverButton("Открыть менеджер проектов", theme.FolderOpenIcon(), func() {
		showProjectList(w)
	})

	// блок авторизации
	var authWidget fyne.CanvasObject
	if currentAuth.IsLoggedIn() {
		userLabel := widget.NewLabel(currentAuth.GetEmail())
		userLabel.Alignment = fyne.TextAlignCenter
		userLabel.TextStyle = fyne.TextStyle{Bold: true}

		logoutBtn := widget.NewButtonWithIcon("Выйти", theme.CancelIcon(), func() {
			clearAuthState()
			showStartScreen(w)
		})
		logoutBtn.Importance = widget.LowImportance

		authWidget = container.NewVBox(
			container.NewCenter(container.NewHBox(widget.NewIcon(theme.AccountIcon()), userLabel)),
			container.NewCenter(logoutBtn),
		)
	} else {
		loginBtn := NewThemedHoverButton("Войти в облако", theme.AccountIcon(), func() {
			showLoginScreen(w)
		})
		authWidget = container.NewVBox(
			widget.NewLabelWithStyle("Синхронизация отключена", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			container.NewCenter(container.NewGridWrap(fyne.NewSize(200, 40), loginBtn)),
		)
	}

	// карточка действий
	v := fyne.CurrentApp().Settings().ThemeVariant()
	cardBg := canvas.NewRectangle(theme.Current().Color(ColorNameCardBackground, v))
	cardBg.CornerRadius = 24
	cardBg.StrokeColor = theme.PrimaryColor()
	cardBg.StrokeWidth = 0.5

	mainActions := container.NewStack(
		cardBg,
		container.NewPadded(container.NewVBox(
			layout.NewSpacer(),
			container.NewCenter(container.NewGridWrap(fyne.NewSize(340, 56), startBtn)),
			authWidget,
			layout.NewSpacer(),
		)),
	)

	// сборка контента
	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(title),
		subtitle,
		widget.NewLabel(""),
		container.NewCenter(container.NewGridWrap(fyne.NewSize(460, 260), mainActions)),
		layout.NewSpacer(),
		layout.NewSpacer(),
	)

	w.SetContent(container.NewStack(
		canvas.NewRectangle(theme.BackgroundColor()),
		content,
	))
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
}

// экран выбора и создания проекта
func showProjectList(w fyne.Window) {
	appData := loadProjects()

	title := widget.NewLabel("Менеджер проектов")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	// панель статистики
	totalProjectsLabel := widget.NewLabel("")

	updateStats := func() {
		totalProjectsLabel.SetText(fmt.Sprintf("Всего проектов: %d", len(appData.Projects)))
	}
	updateStats()

	statsBar := container.NewHBox(
		container.NewPadded(totalProjectsLabel),
		layout.NewSpacer(),
	)

	projectList := container.NewVBox()
	cards := []*projectCard{}

	renderProjects := func(filter string) {
		projectList.RemoveAll()
		cards = []*projectCard{}
		filter = strings.ToLower(filter)

		found := false
		for i := range appData.Projects {
			idx := i
			proj := appData.Projects[idx]

			if filter != "" && !strings.Contains(strings.ToLower(proj.Name), filter) {
				continue
			}
			found = true

			card := createProjectCard(w, proj,
				func() {
					showProject(w, proj.Name)
				},
				func() {
					dialog.ShowConfirm("Удалить проект",
						fmt.Sprintf("Удалить проект «%s»?", proj.Name),
						func(ok bool) {
							if !ok {
								return
							}
							appData.Projects = append(appData.Projects[:idx], appData.Projects[idx+1:]...)
							_ = saveProjects(appData)
							showProjectList(w)
						}, w)
				})
			cards = append(cards, card)
			projectList.Add(container.NewPadded(card.container))
		}

		if !found {
			msg := "Нет проектов"
			if filter != "" {
				msg = "Ничего не найдено"
			}
			projectList.Add(container.NewCenter(container.NewVBox(
				widget.NewLabel(""),
				widget.NewLabel(msg),
			)))
		}
		projectList.Refresh()
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Поиск проекта...")
	searchEntry.OnChanged = renderProjects

	renderProjects("")

	createBtn := widget.NewButtonWithIcon("Создать проект", theme.ContentAddIcon(), func() {
		nameEntry := widget.NewEntry()
		nameEntry.SetPlaceHolder("Название проекта")

		dialog.ShowForm("Новый проект", "Создать", "Отмена",
			[]*widget.FormItem{
				widget.NewFormItem("Имя", nameEntry),
			},
			func(ok bool) {
				if !ok || strings.TrimSpace(nameEntry.Text) == "" {
					return
				}
				newProject := Project{
					Name:      strings.TrimSpace(nameEntry.Text),
					Equipment: []Equipment{},
				}
				appData.Projects = append(appData.Projects, newProject)
				if err := saveProjects(appData); err != nil {
					dialog.ShowError(err, w)
					return
				}
				showProjectList(w)
			}, w)
	})
	createBtn.Importance = widget.HighImportance

	backBtn := widget.NewButtonWithIcon("Назад", theme.NavigateBackIcon(), func() {
		showStartScreen(w)
	})

	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), func() {
		current := fyne.CurrentApp().Settings().Theme()
		if m, ok := current.(*modernTheme); ok && m.variant == theme.VariantDark {
			fyne.CurrentApp().Settings().SetTheme(newModernLightTheme())
		} else {
			fyne.CurrentApp().Settings().SetTheme(newModernDarkTheme())
		}
		// обновление карточек
		for _, c := range cards {
			c.refreshTheme()
		}
		w.Content().Refresh()
	})

	// кнопки облака для авторизованных пользователей
	var cloudButtons fyne.CanvasObject
	if currentAuth.IsLoggedIn() {
		cloudSaveInfo := widget.NewLabel("☁ " + currentAuth.GetEmail())
		cloudSaveInfo.TextStyle = fyne.TextStyle{Italic: true}

		cloudLoadBtn := widget.NewButtonWithIcon("Загрузить из облака", theme.DownloadIcon(), func() {
			showCloudLoadDialog(w)
		})
		cloudButtons = container.NewHBox(cloudSaveInfo, layout.NewSpacer(), cloudLoadBtn)
	}

	scrollable := container.NewVScroll(projectList)
	scrollable.SetMinSize(fyne.NewSize(600, 400))

	// вкладки личных и командных проектов
	var mainContent fyne.CanvasObject
	if currentAuth.IsLoggedIn() {
		myProjectsTab := container.NewTabItem("Мои проекты", container.NewBorder(nil, container.NewPadded(createBtn), nil, nil, scrollable))
		
		teamProjectsBox := container.NewVBox()
		teamScrollable := container.NewVScroll(teamProjectsBox)
		
		teamSelect := widget.NewSelect([]string{}, nil)
		teamMap := make(map[string]CloudTeam)
		
		loadTeamData := func() {
			teams, err := loadTeams(currentAuth.GetToken())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			names := make([]string, 0, len(teams))
			for _, t := range teams {
				names = append(names, t.Name)
				teamMap[t.Name] = t
			}
			teamSelect.Options = names
			teamSelect.Refresh()
			
			// выбор первой команды или обновление проектов выбранной
			if len(names) > 0 && teamSelect.Selected == "" {
				teamSelect.SetSelected(names[0])
			} else if teamSelect.Selected != "" {
				teamSelect.OnChanged(teamSelect.Selected)
			}
		}
		
		teamSelect.OnChanged = func(name string) {
			t, ok := teamMap[name]
			if !ok { return }
			teamProjectsBox.RemoveAll()
			
			projects, err := loadTeamProjects(t.ID, currentAuth.GetToken())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			
			for _, cp := range projects {
				cloudProj := cp // capture
				card := createProjectCard(w, Project{
					Name:             cloudProj.Name, 
					TeamID:           cloudProj.TeamID,
					OwnerID:          cloudProj.OwnerID,
					CloudItemCount:   cloudProj.ItemCount,
					CloudTotalWeight: cloudProj.TotalWeight,
				}, func() {
					// загрузка детальной информации и локальное сохранение
					token := currentAuth.GetToken()
					url := fmt.Sprintf("%s/%d", apiProjectsURL, cloudProj.ID)
					req, _ := http.NewRequest(http.MethodGet, url, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					client := &http.Client{Timeout: 15 * time.Second}
					resp, err := client.Do(req)
					if err != nil || resp.StatusCode != http.StatusOK {
						dialog.ShowError(fmt.Errorf("не удалось загрузить проект"), w)
						return
					}
					defer resp.Body.Close()
					var detail cloudProjectDetail
					if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
						dialog.ShowError(err, w)
						return
					}
					
					localProj := cloudProjectToLocal(&detail)
					// обновление или добавление в локальный список проектов
					found := false
					for i, p := range appData.Projects {
						if p.CloudID == localProj.CloudID || p.Name == localProj.Name {
							appData.Projects[i] = localProj
							found = true
							break
						}
					}
					if !found {
						appData.Projects = append(appData.Projects, localProj)
					}
					_ = saveProjects(appData)
					showProject(w, localProj.Name)
				}, func() {})
				// удаление отключено для не-владельцев
				teamProjectsBox.Add(container.NewPadded(card.container))
			}
			teamProjectsBox.Refresh()
		}
		
		manageTeamsBtn := widget.NewButtonWithIcon("Управление командами", theme.SettingsIcon(), func() {
			showTeamManagement(w)
		})
		manageTeamsBtn.Importance = widget.LowImportance

		teamTopBar := container.NewHBox(widget.NewLabel("Команда:"), teamSelect, layout.NewSpacer(), manageTeamsBtn)
		
		createTeamProjectBtn := widget.NewButtonWithIcon("Создать проект команды", theme.ContentAddIcon(), func() {
			if teamSelect.Selected == "" {
				dialog.ShowInformation("Внимание", "Сначала выберите команду", w)
				return
			}
			t, ok := teamMap[teamSelect.Selected]
			if !ok { return }

			nameEntry := widget.NewEntry()
			nameEntry.SetPlaceHolder("Название проекта")

			dialog.ShowForm("Новый проект команды", "Создать", "Отмена",
				[]*widget.FormItem{
					widget.NewFormItem("Имя", nameEntry),
				},
				func(ok bool) {
					if !ok || strings.TrimSpace(nameEntry.Text) == "" {
						return
					}
					newProject := Project{
						Name:      strings.TrimSpace(nameEntry.Text),
						Equipment: []Equipment{},
						TeamID:    &t.ID,
					}
					
					go func() {
						cloudID, err := saveProjectToCloud(newProject, currentAuth.GetToken())
						fyne.Do(func() {
							if err != nil {
								dialog.ShowError(fmt.Errorf("ошибка создания в облаке: %w", err), w)
								return
							}
							newProject.CloudID = cloudID
							appData.Projects = append(appData.Projects, newProject)
							_ = saveProjects(appData)
							
							// открытие нового проекта
							showProject(w, newProject.Name)
						})
					}()
				}, w)
		})
		createTeamProjectBtn.Importance = widget.HighImportance

		teamTabContent := container.NewBorder(container.NewPadded(teamTopBar), container.NewPadded(createTeamProjectBtn), nil, nil, teamScrollable)
		teamProjectsTab := container.NewTabItem("Проекты команд", teamTabContent)
		
		tabs := container.NewAppTabs(myProjectsTab, teamProjectsTab)
		tabs.OnSelected = func(item *container.TabItem) {
			if item == teamProjectsTab {
				loadTeamData()
			}
		}
		mainContent = tabs
	} else {
		mainContent = container.NewBorder(nil, container.NewPadded(createBtn), nil, nil, scrollable)
	}

	headerItems := []fyne.CanvasObject{
		container.NewHBox(backBtn, layout.NewSpacer(), themeBtn),
		title,
		statsBar,
		container.NewPadded(searchEntry),
	}
	if cloudButtons != nil {
		headerItems = append(headerItems, container.NewPadded(cloudButtons))
	}
	headerItems = append(headerItems, widget.NewSeparator())
	header := container.NewVBox(headerItems...)

	content := container.NewBorder(
		header,
		nil,
		nil, nil,
		mainContent,
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
}

// экран управления командами
func showTeamManagement(w fyne.Window) {
	title := widget.NewLabel("Управление командами")
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	var teamList *widget.List
	var memberBox *fyne.Container
	var teams []CloudTeam
	var currentTeam *CloudTeam

	loadTeamsData := func() {
		t, err := loadTeams(currentAuth.GetToken())
		if err == nil {
			teams = t
			teamList.Refresh()
		}
	}

	var loadMembersData func(team CloudTeam)
	loadMembersData = func(team CloudTeam) {
		members, err := loadTeamMembers(team.ID, currentAuth.GetToken())
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		memberBox.RemoveAll()
		
		isOwner := team.OwnerID == currentAuth.UserID
		
		for _, m := range members {
			usr := m // capture
			lbl := widget.NewLabel(fmt.Sprintf("%s (%s)", usr.Email, usr.Role))
			var actBtn *widget.Button
			if isOwner && usr.UserID != currentAuth.UserID {
				actBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					// вызов api удаления участника
					url := fmt.Sprintf("%s/%d/members?user_id=%d", apiTeamsURL, team.ID, usr.UserID)
					req, _ := http.NewRequest(http.MethodDelete, url, nil)
					req.Header.Set("Authorization", "Bearer "+currentAuth.GetToken())
					client := &http.Client{Timeout: 15 * time.Second}
					resp, err := client.Do(req)
					if err == nil && resp.StatusCode == http.StatusOK {
						loadMembersData(team)
					} else {
						dialog.ShowError(fmt.Errorf("ошибка удаления участника"), w)
					}
				})
			} else if !isOwner && usr.UserID == currentAuth.UserID {
				// выход участника из команды
				actBtn = widget.NewButton("Выйти", func() {
					url := fmt.Sprintf("%s/%d/members?user_id=%d", apiTeamsURL, team.ID, usr.UserID)
					req, _ := http.NewRequest(http.MethodDelete, url, nil)
					req.Header.Set("Authorization", "Bearer "+currentAuth.GetToken())
					client := &http.Client{Timeout: 15 * time.Second}
					resp, err := client.Do(req)
					if err == nil && resp.StatusCode == http.StatusOK {
						loadTeamsData()
						memberBox.RemoveAll()
					}
				})
			} else {
				actBtn = widget.NewButton("", nil)
				actBtn.Hide()
			}
			row := container.NewHBox(lbl, layout.NewSpacer(), actBtn)
			memberBox.Add(row)
		}
		
		if isOwner {
			addBtn := widget.NewButtonWithIcon("Добавить участника", theme.ContentAddIcon(), func() {
				emailEntry := widget.NewEntry()
				emailEntry.SetPlaceHolder("Email пользователя")
				dialog.ShowForm("Добавить участника", "Добавить", "Отмена",
					[]*widget.FormItem{widget.NewFormItem("Email", emailEntry)},
					func(ok bool) {
						if ok && emailEntry.Text != "" {
							if err := addTeamMember(team.ID, emailEntry.Text, currentAuth.GetToken()); err != nil {
								dialog.ShowError(err, w)
							} else {
								loadMembersData(team)
							}
						}
					}, w)
			})
			memberBox.Add(container.NewPadded(addBtn))
		}
		memberBox.Refresh()
	}

	teamList = widget.NewList(
		func() int { return len(teams) },
		func() fyne.CanvasObject { return widget.NewLabel("Team Name") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(teams[i].Name)
		},
	)
	teamList.OnSelected = func(id widget.ListItemID) {
		currentTeam = &teams[id]
		loadMembersData(*currentTeam)
	}

	memberBox = container.NewVBox()
	
	createTeamBtn := widget.NewButtonWithIcon("Создать команду", theme.ContentAddIcon(), func() {
		nameEntry := widget.NewEntry()
		dialog.ShowForm("Новая команда", "Создать", "Отмена",
			[]*widget.FormItem{widget.NewFormItem("Название", nameEntry)},
			func(ok bool) {
				if ok && nameEntry.Text != "" {
					if _, err := createTeam(nameEntry.Text, currentAuth.GetToken()); err != nil {
						dialog.ShowError(err, w)
					} else {
						loadTeamsData()
					}
				}
			}, w)
	})

	backBtn := widget.NewButtonWithIcon("Назад", theme.NavigateBackIcon(), func() {
		showProjectList(w)
	})

	topBar := container.NewHBox(backBtn, layout.NewSpacer(), title, layout.NewSpacer())

	leftPanel := container.NewBorder(container.NewPadded(createTeamBtn), nil, nil, nil, teamList)
	rightPanel := container.NewVScroll(memberBox)
	
	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.4

	content := container.NewBorder(topBar, nil, nil, nil, split)
	w.SetContent(container.NewPadded(content))
	
	loadTeamsData()
}

// главное окно проекта со списком оборудования
func showProject(w fyne.Window, projectName string) {
	appData := loadProjects()

	var projIdx int = -1
	for i, p := range appData.Projects {
		if p.Name == projectName {
			projIdx = i
			break
		}
	}
	if projIdx == -1 {
		dialog.ShowError(fmt.Errorf("проект '%s' не найден", projectName), w)
		return
	}

	title := widget.NewLabel(fmt.Sprintf("Проект: %s", projectName))
	title.Alignment = fyne.TextAlignCenter
	title.TextStyle = fyne.TextStyle{Bold: true}

	rowsContainer := container.NewVBox()
	var rows []*equipmentRow

	totalWeightLabel := widget.NewLabel("—")
	totalWeightLabel.TextStyle = fyne.TextStyle{Bold: true}
	totalWeightLabel.Alignment = fyne.TextAlignLeading
	byTypeLabel := widget.NewLabel("По типам: —")

	recalcAll := func() {
		// обновление итоговых весов в главном потоке
		fyne.Do(func() {
			var grandTotal float64
			typeWeights := make(map[string]float64)

			for _, r := range rows {
				qty, _ := strconv.Atoi(strings.TrimSpace(r.qtyEntry.Text))
				if qty < 1 {
					qty = 1
				}
				eqType := r.typeSelect.Selected

				var unitWeight float64
				text := r.resultLabel.Text
				if strings.HasSuffix(text, " кг/ед.") {
					text = strings.TrimSuffix(text, " кг/ед.")
					text = strings.TrimPrefix(text, "✓ ")
					text = strings.TrimSpace(text)
					unitWeight, _ = strconv.ParseFloat(text, 64)
				}

				lineTotal := unitWeight * float64(qty)
				grandTotal += lineTotal
				if unitWeight > 0 {
					typeWeights[eqType] += lineTotal
				}
			}

			totalWeightLabel.SetText(fmt.Sprintf("%.2f кг", grandTotal))

			var parts []string
			for _, t := range equipmentTypes {
				if ww, ok := typeWeights[t]; ok && ww > 0 {
					parts = append(parts, fmt.Sprintf("%s: %.2f кг", t, ww))
				}
			}
			if len(parts) > 0 {
				byTypeLabel.SetText("По типам: " + strings.Join(parts, " | "))
			} else {
				byTypeLabel.SetText("По типам: —")
			}
		})
	}

	removeRow := func(target *equipmentRow) {
		newRows := make([]*equipmentRow, 0, len(rows))
		for _, r := range rows {
			if r != target {
				newRows = append(newRows, r)
			}
		}
		rows = newRows
		rowsContainer.Remove(target.container)
		rowsContainer.Refresh()
		recalcAll()
	}

	buildFieldsByType := func(row *equipmentRow, eqType string) *fyne.Container {
		switch eqType {
		case "Насосы":
			return buildPumpFields(row)
		case "Конвейер":
			return buildConveyorFields(row)
		case "Вертикальный аппарат":
			return buildVesselFields(row)
		case "Горизонтальная емкость":
			return buildDrumFields(row)
		case "Трубчатый теплообменник":
			return buildUTubeFields(row)
		case "Тарельчатая колонна":
			return buildTowerFields(row)
		case "Коробчатая технологическая печь":
			return buildBoxFurnaceFields(row)
		case "Центробежный компрессор":
			return buildCentrifugalCompressorFields(row)
		default:
			return container.NewVBox(widget.NewLabel("Неизвестный тип оборудования"))
		}
	}

	addEquipmentRow := func(eq Equipment) {
		row := &equipmentRow{loading: true}

		row.typeSelect = widget.NewSelect(equipmentTypes, nil)
		if eq.Type != "" {
			row.typeSelect.SetSelected(eq.Type)
		} else {
			row.typeSelect.SetSelectedIndex(0)
		}

		row.tagEntry = widget.NewEntry()
		row.tagEntry.SetPlaceHolder("Тэг / Имя")
		row.tagEntry.SetText(eq.Tag)
		row.tagEntry.OnChanged = func(s string) {
			row.markFieldInvalid(row.tagEntry, nil, strings.TrimSpace(s) == "")
			row.dataChanged(false)
		}
		tagContainer := container.NewGridWrap(fyne.NewSize(150, 40), row.tagEntry)

		row.qtyEntry = widget.NewEntry()
		row.qtyEntry.SetPlaceHolder("Кол-во")
		if eq.Quantity > 0 {
			row.qtyEntry.SetText(strconv.Itoa(eq.Quantity))
		} else {
			row.qtyEntry.SetText("1")
		}
		row.qtyEntry.OnChanged = func(s string) {
			q, err := strconv.Atoi(strings.TrimSpace(s))
			row.markFieldInvalid(row.qtyEntry, nil, err != nil || q < 1)
			row.dataChanged(false)
		}

		row.resultLabel = widget.NewLabel("—")
		row.resultLabel.TextStyle = fyne.TextStyle{Bold: true}

		if eq.CalculatedWeight > 0 {
			row.resultLabel.SetText(fmt.Sprintf("%.2f кг/ед.", eq.CalculatedWeight))
		}

		currentType := row.typeSelect.Selected
		row.fieldsContainer = buildFieldsByType(row, currentType)
		qtyContainer := container.NewGridWrap(fyne.NewSize(70, 40), row.qtyEntry)

		switch currentType {
		case "Насосы":
			row.flowEntry.SetText(floatPtrToStr(eq.FlowRate))
			row.headEntry.SetText(floatPtrToStr(eq.FluidHead))
			row.rpmEntry.SetText(floatPtrToStr(eq.RPM))
			row.specGravityEntry.SetText(floatPtrToStr(eq.SpecGravity))
			row.powerEntry.SetText(floatPtrToStr(eq.PowerKW))
		case "Конвейер":
			row.conveyorLengthEntry.SetText(floatPtrToStr(eq.ConveyorLength))
			row.beltWidthEntry.SetText(floatPtrToStr(eq.BeltWidth))
			row.conveyorFlowRateEntry.SetText(floatPtrToStr(eq.ConveyorFlowRate))
		case "Вертикальный аппарат":
			row.vesselDiameterEntry.SetText(floatPtrToStr(eq.VesselDiameter))
			row.vesselTangentToTangentHeightEntry.SetText(floatPtrToStr(eq.VesselTangentToTangentHeight))
			row.designGaugePressureEntry.SetText(floatPtrToStr(eq.DesignGaugePressure))
			row.designTemperatureEntry.SetText(floatPtrToStr(eq.DesignTemperature))
			row.skirtHeightEntry.SetText(floatPtrToStr(eq.SkirtHeight))
			row.vesselLegHeightEntry.SetText(floatPtrToStr(eq.VesselLegHeight))
		case "Горизонтальная емкость":
			row.vesselDiameterEntry.SetText(floatPtrToStr(eq.VesselDiameter))
			row.designTangentToTangentLengthEntry.SetText(floatPtrToStr(eq.DesignTangentToTangentLength))
			row.designGaugePressureEntry.SetText(floatPtrToStr(eq.DesignGaugePressure))
			row.designTemperatureEntry.SetText(floatPtrToStr(eq.DesignTemperature))
		case "Трубчатый теплообменник":
			row.shellDiameterEntry.SetText(floatPtrToStr(eq.ShellDiameter))
			row.tubeOutDiameterEntry.SetText(floatPtrToStr(eq.TubeOutDiameter))
			row.tubeLenEntry.SetText(floatPtrToStr(eq.TubeLen))
			row.tubeDesPresEntry.SetText(floatPtrToStr(eq.TubeDesPres))
			row.heatAreaEntry.SetText(floatPtrToStr(eq.HeatArea))
		case "Тарельчатая колонна":
			row.vesselDiameterEntry.SetText(floatPtrToStr(eq.VesselDiameter))
			row.numberOfTraysEntry.SetText(floatPtrToStr(eq.NumberOfTrays))
			row.designTangentToTangentLengthEntry.SetText(floatPtrToStr(eq.DesignTangentToTangentLength))
			row.designGaugePressureEntry.SetText(floatPtrToStr(eq.DesignGaugePressure))
		case "Коробчатая технологическая печь":
			row.dutyEntry.SetText(floatPtrToStr(eq.Duty))
			row.standardGasFlowRateEntry.SetText(floatPtrToStr(eq.StandardGasFlowRate))
			row.designGaugePressureEntry.SetText(floatPtrToStr(eq.DesignGaugePressure))
			row.designTemperatureEntry.SetText(floatPtrToStr(eq.DesignTemperature))
		case "Центробежный компрессор":
			row.actualGasFlowRateInletEntry.SetText(floatPtrToStr(eq.ActualGasFlowRateInlet))
			row.designGaugePressureInletEntry.SetText(floatPtrToStr(eq.DesignGaugePressureInlet))
			row.designGaugePressureOutletEntry.SetText(floatPtrToStr(eq.DesignGaugePressureOutlet))
			row.driverPowerEntry.SetText(floatPtrToStr(eq.DriverPower))
		}

		calcBtn := widget.NewButtonWithIcon("Рассчитать", theme.ConfirmIcon(), func() {
			eqData, err := row.collectEquipment()
			if err != nil {
				row.resultLabel.SetText("⚠ " + err.Error())
				return
			}

			fyne.Do(func() {
				row.resultLabel.SetText("⏳ Расчёт...")
				row.resultLabel.Refresh()
			})

			go func() {
				weight, err := sendEquipmentToBackend(eqData)
				fyne.Do(func() {
					if err != nil {
						row.resultLabel.SetText("✗ " + err.Error())
					} else {
						row.resultLabel.SetText(fmt.Sprintf("%.2f кг/ед.", weight))
					}
					row.resultLabel.Refresh()
					recalcAll()
				})
			}()
		})
		calcBtn.Importance = widget.HighImportance

		row.deleteBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			removeRow(row)
		})
		row.deleteBtn.Importance = widget.LowImportance

		row.deleteBg = canvas.NewRectangle(theme.InputBackgroundColor())
		row.deleteBg.CornerRadius = theme.InputRadiusSize()
		styledDeleteBtn := container.NewStack(row.deleteBg, row.deleteBtn)

		row.typeSelect.OnChanged = func(selected string) {
			newFields := buildFieldsByType(row, selected)

			row.fieldsContainer.RemoveAll()
			for _, obj := range newFields.Objects {
				row.fieldsContainer.Add(obj)
			}
			row.fieldsContainer.Show()
			row.expandBtn.SetIcon(theme.MenuDropUpIcon())
			row.fieldsContainer.Refresh()
			row.dataChanged(true)
		}

		row.expandBtn = widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), func() {
			if row.fieldsContainer.Visible() {
				row.fieldsContainer.Hide()
				row.expandBtn.SetIcon(theme.MenuDropDownIcon())
			} else {
				row.fieldsContainer.Show()
				row.expandBtn.SetIcon(theme.MenuDropUpIcon())
			}
			row.container.Refresh()
		})
		row.expandBtn.Importance = widget.LowImportance

		// кнопка сворачивания карточки
		row.expandBg = canvas.NewRectangle(theme.InputBackgroundColor())
		row.expandBg.CornerRadius = theme.InputRadiusSize()
		styledExpandBtn := container.NewGridWrap(fyne.NewSize(40, 40), container.NewStack(row.expandBg, row.expandBtn))

		// карточный заголовок
		weightHeaderLabel := widget.NewLabelWithStyle("Вес единицы:", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})

		topRow := container.NewHBox(
			styledExpandBtn,
			container.NewGridWrap(fyne.NewSize(180, 40), row.typeSelect),
			tagContainer,
			container.NewHBox(widget.NewLabel("Кол-во:"), qtyContainer),
			layout.NewSpacer(),
			container.NewHBox(
				weightHeaderLabel,
				row.resultLabel,
			),
			container.NewGridWrap(fyne.NewSize(140, 40), calcBtn),
			container.NewGridWrap(fyne.NewSize(40, 40), styledDeleteBtn),
		)

		// настройка внешнего вида результата
		row.resultLabel.Alignment = fyne.TextAlignTrailing
		row.resultLabel.TextStyle = fyne.TextStyle{Bold: true}

		// упаковка в карточку с фоном и акцентом
		v := fyne.CurrentApp().Settings().ThemeVariant()
		row.cardBg = canvas.NewRectangle(theme.Current().Color(ColorNameCardBackground, v))
		row.cardBg.CornerRadius = 12

		row.accentBar = canvas.NewRectangle(theme.PrimaryColor())
		row.accentBar.SetMinSize(fyne.NewSize(4, 0))

		content := container.NewPadded(container.NewVBox(
			topRow,
			container.NewPadded(row.fieldsContainer),
		))

		cardContent := container.NewHBox(row.accentBar, content)

		row.container = container.NewStack(
			row.cardBg,
			cardContent,
		)

		// отступы между карточками
		row.container = container.NewPadded(row.container)

		rows = append(rows, row)
		rowsContainer.Add(row.container)
		rowsContainer.Refresh()

		row.loading = false
		// установка callback после инициализации полей
		row.OnChanged = recalcAll
	}

	for _, eq := range appData.Projects[projIdx].Equipment {
		addEquipmentRow(eq)
	}
	recalcAll()

	addBtn := widget.NewButtonWithIcon("Добавить единицу оборудования", theme.ContentAddIcon(), func() {
		addEquipmentRow(Equipment{})
	})
	addBtn.Importance = widget.HighImportance

	calcAllBtn := widget.NewButtonWithIcon("Рассчитать всё", theme.ComputerIcon(), func() {
		fyne.Do(func() {
			totalWeightLabel.SetText("⏳ Расчёт...")
		})

		go func() {
			for _, r := range rows {
				var eq Equipment
				var err error
				fyne.DoAndWait(func() {
					eq, err = r.collectEquipment()
					if err == nil {
						r.resultLabel.SetText("⏳...")
						r.resultLabel.Refresh()
					}
				})
				if err != nil {
					continue
				}

				weight, err := sendEquipmentToBackend(eq)
				fyne.Do(func() {
					if err != nil {
						r.resultLabel.SetText("✗ " + err.Error())
					} else {
						r.resultLabel.SetText(fmt.Sprintf("%.2f кг/ед.", weight))
					}
					r.resultLabel.Refresh()
				})
			}
			recalcAll()
		}()
	})
	calcAllBtn.Importance = widget.HighImportance

	saveBtn := widget.NewButtonWithIcon("Сохранить проект", theme.DocumentSaveIcon(), func() {
		var equipment []Equipment
		hasUncalculated := false
		for _, r := range rows {
			eq, err := r.collectEquipment()
			if err != nil {
				continue
			}
			text := r.resultLabel.Text
			if strings.HasSuffix(text, " кг/ед.") {
				text = strings.TrimSuffix(text, " кг/ед.")
				text = strings.TrimPrefix(text, "✓ ")
				eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
			} else {
				hasUncalculated = true
			}
			equipment = append(equipment, eq)
		}

		doLocalSave := func(eqList []Equipment) {
			appData.Projects[projIdx].Equipment = eqList
			appData.Projects[projIdx].UpdatedAt = time.Now()
			if err := saveProjects(appData); err != nil {
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("Сохранено", "Проект успешно сохранён.", w)
		}

		if hasUncalculated {
			continueBtn := widget.NewButton("Всё равно сохранить", func() {
				doLocalSave(equipment)
			})
			var warningDialog dialog.Dialog
			calcAndSaveBtn := widget.NewButton("Рассчитать и сохранить", func() {
				if warningDialog != nil {
					warningDialog.Hide()
				}
				fyne.Do(func() {
					totalWeightLabel.SetText("⏳ Расчёт...")
				})
				go func() {
					for _, r := range rows {
						var eq Equipment
						var err error
						fyne.DoAndWait(func() {
							eq, err = r.collectEquipment()
							if err == nil {
								r.resultLabel.SetText("⏳...")
								r.resultLabel.Refresh()
							}
						})
						if err != nil {
							continue
						}

						weight, err := sendEquipmentToBackend(eq)
						fyne.Do(func() {
							if err != nil {
								r.resultLabel.SetText("✗ " + err.Error())
							} else {
								r.resultLabel.SetText(fmt.Sprintf("✓ %.2f кг/ед.", weight))
							}
							r.resultLabel.Refresh()
						})
					}
					recalcAll()

					// сбор данных оборудования после расчета
					var updatedEquipment []Equipment
					for _, r := range rows {
						eq, _ := r.collectEquipment()
						text := r.resultLabel.Text
						if strings.HasSuffix(text, " кг/ед.") {
							text = strings.TrimSuffix(text, " кг/ед.")
							text = strings.TrimPrefix(text, "✓ ")
							eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
						}
						updatedEquipment = append(updatedEquipment, eq)
					}
					doLocalSave(updatedEquipment)
				}()
			})
			calcAndSaveBtn.Importance = widget.HighImportance

			warningContent := container.NewVBox(
				widget.NewLabel("Не у всех строк рассчитан вес.\nСохранить проект без этих данных?"),
				container.NewHBox(continueBtn, calcAndSaveBtn),
			)
			warningDialog = dialog.NewCustomWithoutButtons("Предупреждение", warningContent, w)
			continueBtn.OnTapped = func() {
				warningDialog.Hide()
				doLocalSave(equipment)
			}
			warningDialog.Show()
		} else {
			doLocalSave(equipment)
		}
	})

	backBtn := widget.NewButtonWithIcon("Назад к проектам", theme.NavigateBackIcon(), func() {
		showProjectList(w)
	})

	scrollable := container.NewVScroll(rowsContainer)
	scrollable.SetMinSize(fyne.NewSize(800, 400))

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabelWithStyle("Итоговый вес проекта:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			totalWeightLabel,
		),
		byTypeLabel,
	)

	isAllExpanded := true
	toggleAllBtn := widget.NewButtonWithIcon("Свернуть всё", theme.MenuDropUpIcon(), nil)
	toggleAllBtn.OnTapped = func() {
		if isAllExpanded {
			for _, r := range rows {
				r.fieldsContainer.Hide()
				r.expandBtn.SetIcon(theme.MenuDropDownIcon())
				r.container.Refresh()
			}
			toggleAllBtn.SetText("Развернуть всё")
			toggleAllBtn.SetIcon(theme.MenuDropDownIcon())
			isAllExpanded = false
		} else {
			for _, r := range rows {
				r.fieldsContainer.Show()
				r.expandBtn.SetIcon(theme.MenuDropUpIcon())
				r.container.Refresh()
			}
			toggleAllBtn.SetText("Свернуть всё")
			toggleAllBtn.SetIcon(theme.MenuDropUpIcon())
			isAllExpanded = true
		}
	}

	// кнопка справки
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		instructionText := `Инструкция по работе с шаблоном Excel:

1. Нажмите «Шаблон» для скачивания пустого файла.
2. Заполните соответствующие листы шаблона.
3. Не меняйте порядок колонок.
4. Обязательные поля отмечены в шаблоне красным цветом.
5. Нажмите «Импорт» и выберите заполненный файл.

Поддерживаемые типы оборудования:
  • Насосы
  • Конвейер
  • Вертикальные аппараты
  • Горизонтальные емкости

При импорте данные добавляются к текущему проекту.`
		dialog.ShowInformation("Справка: Импорт/Экспорт", instructionText, w)
	})

	// кнопка скачивания шаблона
	templateBtn := widget.NewButtonWithIcon("Шаблон", theme.DownloadIcon(), func() {
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if writer == nil {
				return // пользователь отменил
			}
			filePath := writer.URI().Path()
			writer.Close()
			// удаление временного файла перед записью excelize
			os.Remove(filePath)

			if err := generateTemplate(filePath); err != nil {
				dialog.ShowError(fmt.Errorf("ошибка создания шаблона: %w", err), w)
				return
			}
			dialog.ShowInformation("Готово", "Шаблон успешно сохранён.", w)
		}, w)
		saveDialog.SetFileName("шаблон_оборудования.xlsx")
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
		saveDialog.Show()
	})

	// кнопка импорта
	importBtn := widget.NewButtonWithIcon("Импорт", theme.FolderOpenIcon(), func() {
		openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return // пользователь отменил
			}
			filePath := reader.URI().Path()
			reader.Close()

			imported, importErrors := importProject(filePath)

			// добавление импортированного оборудования
			for _, eq := range imported {
				addEquipmentRow(eq)
			}
			recalcAll()

			// отображение отчета об импорте
			if len(importErrors) > 0 {
				var report strings.Builder
				report.WriteString(fmt.Sprintf("Успешно импортировано: %d записей.\n\n", len(imported)))
				report.WriteString("Ошибки при импорте:\n")
				for _, ie := range importErrors {
					report.WriteString("• " + ie.String() + "\n")
				}
				dialog.ShowInformation("Результат импорта", report.String(), w)
			} else if len(imported) > 0 {
				dialog.ShowInformation("Импорт завершён",
					fmt.Sprintf("Успешно импортировано: %d записей.", len(imported)), w)
			} else {
				dialog.ShowInformation("Импорт", "Файл не содержит данных для импорта.", w)
			}
		}, w)
		openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
		openDialog.Show()
	})

	// кнопка экспорта
	exportBtn := widget.NewButtonWithIcon("Экспорт", theme.DocumentCreateIcon(), func() {
		// сбор оборудования из строк
		var equipment []Equipment
		hasUncalculated := false
		for _, r := range rows {
			eq, err := r.collectEquipment()
			if err != nil {
				continue
			}
			text := r.resultLabel.Text
			if strings.HasSuffix(text, " кг/ед.") {
				text = strings.TrimSuffix(text, " кг/ед.")
				text = strings.TrimPrefix(text, "✓ ")
				eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
			} else {
				hasUncalculated = true
			}
			equipment = append(equipment, eq)
		}

		if len(equipment) == 0 {
			dialog.ShowInformation("Экспорт", "Нет данных для экспорта.", w)
			return
		}

		doExport := func(eqList []Equipment) {
			saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				if writer == nil {
					return
				}
				filePath := writer.URI().Path()
				writer.Close()
				os.Remove(filePath)

				if err := exportProject(filePath, eqList); err != nil {
					dialog.ShowError(fmt.Errorf("ошибка экспорта: %w", err), w)
					return
				}
				dialog.ShowInformation("Готово", "Данные успешно экспортированы.", w)
			}, w)
			saveDialog.SetFileName(projectName + ".xlsx")
			saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xlsx"}))
			saveDialog.Show()
		}

		if hasUncalculated {
			// диалог подтверждения сохранения нерассчитанных строк
			continueBtn := widget.NewButton("Всё равно продолжить", func() {
				// экспорт без нерассчитанных строк
				doExport(equipment)
			})
			calcBtn := widget.NewButton("Провести расчёт", func() {
				// запуск расчета всех строк
				fyne.Do(func() {
					totalWeightLabel.SetText("⏳ Расчёт...")
				})
				go func() {
					for _, r := range rows {
						var eq Equipment
						var err error
						fyne.DoAndWait(func() {
							eq, err = r.collectEquipment()
							if err == nil {
								r.resultLabel.SetText("⏳...")
								r.resultLabel.Refresh()
							}
						})
						if err != nil {
							continue
						}

						weight, err := sendEquipmentToBackend(eq)
						fyne.Do(func() {
							if err != nil {
								r.resultLabel.SetText("✗ " + err.Error())
							} else {
								r.resultLabel.SetText(fmt.Sprintf("✓ %.2f кг/ед.", weight))
							}
							r.resultLabel.Refresh()
						})
					}
					recalcAll()

					// сбор данных после расчета и запуск экспорта
					var updatedEquipment []Equipment
					for _, r := range rows {
						eq, err := r.collectEquipment()
						if err != nil {
							continue
						}
						text := r.resultLabel.Text
						if strings.HasSuffix(text, " кг/ед.") {
							text = strings.TrimSuffix(text, " кг/ед.")
							text = strings.TrimPrefix(text, "✓ ")
							eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
						}
						updatedEquipment = append(updatedEquipment, eq)
					}
					doExport(updatedEquipment)
				}()
			})
			calcBtn.Importance = widget.HighImportance

			warningContent := container.NewVBox(
				widget.NewLabel("Не у всех строк рассчитан вес.\nКолонка «Вес» будет пустой для нерассчитанных строк."),
				container.NewHBox(continueBtn, calcBtn),
			)

			warningDialog := dialog.NewCustomWithoutButtons("Предупреждение", warningContent, w)
			// обработчики кнопок диалога
			continueBtn.OnTapped = func() {
				warningDialog.Hide()
				doExport(equipment)
			}
			calcBtn.OnTapped = func() {
				warningDialog.Hide()
				fyne.Do(func() {
					totalWeightLabel.SetText("⏳ Расчёт...")
				})
				go func() {
					for _, r := range rows {
						var eq Equipment
						var err error
						fyne.DoAndWait(func() {
							eq, err = r.collectEquipment()
							if err == nil {
								r.resultLabel.SetText("⏳...")
								r.resultLabel.Refresh()
							}
						})
						if err != nil {
							continue
						}

						weight, err := sendEquipmentToBackend(eq)
						fyne.Do(func() {
							if err != nil {
								r.resultLabel.SetText("✗ " + err.Error())
							} else {
								r.resultLabel.SetText(fmt.Sprintf("%.2f кг/ед.", weight))
							}
							r.resultLabel.Refresh()
						})
					}
					recalcAll()

					var updatedEquipment []Equipment
					for _, r := range rows {
						eq, err := r.collectEquipment()
						if err != nil {
							continue
						}
						text := r.resultLabel.Text
						if strings.HasSuffix(text, " кг/ед.") {
							text = strings.TrimSuffix(text, " кг/ед.")
							text = strings.TrimPrefix(text, "✓ ")
							eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
						}
						updatedEquipment = append(updatedEquipment, eq)
					}
					doExport(updatedEquipment)
				}()
			}
			warningDialog.Show()
		} else {
			doExport(equipment)
		}
	})

	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), func() {
		current := fyne.CurrentApp().Settings().Theme()
		if m, ok := current.(*modernTheme); ok && m.variant == theme.VariantDark {
			fyne.CurrentApp().Settings().SetTheme(newModernLightTheme())
		} else {
			fyne.CurrentApp().Settings().SetTheme(newModernDarkTheme())
		}

		// обновление строк оборудования
		for _, r := range rows {
			r.refreshTheme()
		}
		w.Content().Refresh()
	})

	toolbarTop := container.NewHBox(
		backBtn,
		layout.NewSpacer(),
		themeBtn,
		helpBtn,
	)

	toolbarActions := container.NewHBox(
		templateBtn,
		importBtn,
		exportBtn,
		widget.NewSeparator(),
		toggleAllBtn,
		layout.NewSpacer(),
		addBtn,
	)

	bottomButtons := container.NewHBox(
		calcAllBtn,
		saveBtn,
	)

	// сохранение в облако для авторизованных пользователей
	if currentAuth.IsLoggedIn() {
		cloudSaveBtn := widget.NewButtonWithIcon("Сохранить в облако", theme.UploadIcon(), func() {
			var equipment []Equipment
			hasUncalculated := false
			for _, r := range rows {
				eq, err := r.collectEquipment()
				if err != nil {
					continue
				}
				text := r.resultLabel.Text
				if strings.HasSuffix(text, " кг/ед.") {
					text = strings.TrimSuffix(text, " кг/ед.")
					text = strings.TrimPrefix(text, "✓ ")
					eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
				} else {
					hasUncalculated = true
				}
				equipment = append(equipment, eq)
			}

			doCloudSave := func(eqList []Equipment) {
				// обновление локального состояния перед отправкой в облако
				appData.Projects[projIdx].Equipment = eqList
				appData.Projects[projIdx].UpdatedAt = time.Now()
				
				currentProject := appData.Projects[projIdx]
				token := currentAuth.GetToken()
				go func() {
					cloudID, err := saveProjectToCloud(currentProject, token)
					fyne.Do(func() {
						if err != nil {
							if err.Error() == "CONFLICT_VERSION" {
								dialog.ShowConfirm("Конфликт версий",
									"Проект был изменён другим участником команды.\nПерезаписать текущей версией?",
									func(ok bool) {
										if ok {
											dialog.ShowInformation("Конфликт", "Для перезаписи обновите проект и повторите.", w)
										}
									}, w)
							} else {
								dialog.ShowError(fmt.Errorf("ошибка облачного сохранения: %w", err), w)
							}
						} else {
							// сохранение изменений и cloudID локально
							if cloudID > 0 {
								appData.Projects[projIdx].CloudID = cloudID
							}
							_ = saveProjects(appData)
							
							dialog.ShowInformation("Облако", "Проект успешно сохранён в облако!", w)
							w.Content().Refresh()
						}
					})
				}()
			}

			if hasUncalculated {
				continueBtn := widget.NewButton("Всё равно сохранить", func() {
					doCloudSave(equipment)
				})
				var warningDialog dialog.Dialog
				calcAndSaveBtn := widget.NewButton("Рассчитать и сохранить", func() {
					if warningDialog != nil {
						warningDialog.Hide()
					}
					fyne.Do(func() {
						totalWeightLabel.SetText("⏳ Расчёт...")
					})
					go func() {
						for _, r := range rows {
							var eq Equipment
							var err error
							fyne.DoAndWait(func() {
								eq, err = r.collectEquipment()
								if err == nil {
									r.resultLabel.SetText("⏳...")
									r.resultLabel.Refresh()
								}
							})
							if err != nil {
								continue
							}

							weight, err := sendEquipmentToBackend(eq)
							fyne.Do(func() {
								if err != nil {
									r.resultLabel.SetText("✗ " + err.Error())
								} else {
									r.resultLabel.SetText(fmt.Sprintf("✓ %.2f кг/ед.", weight))
								}
								r.resultLabel.Refresh()
							})
						}
						recalcAll()

						var updatedEquipment []Equipment
						for _, r := range rows {
							eq, _ := r.collectEquipment()
							text := r.resultLabel.Text
							if strings.HasSuffix(text, " кг/ед.") {
								text = strings.TrimSuffix(text, " кг/ед.")
								text = strings.TrimPrefix(text, "✓ ")
								eq.CalculatedWeight, _ = strconv.ParseFloat(text, 64)
							}
							updatedEquipment = append(updatedEquipment, eq)
						}
						doCloudSave(updatedEquipment)
					}()
				})
				calcAndSaveBtn.Importance = widget.HighImportance

				warningContent := container.NewVBox(
					widget.NewLabel("Не у всех строк рассчитан вес.\nСохранить в облако без этих данных?"),
					container.NewHBox(continueBtn, calcAndSaveBtn),
				)
				warningDialog = dialog.NewCustomWithoutButtons("Предупреждение", warningContent, w)
				continueBtn.OnTapped = func() {
					warningDialog.Hide()
					doCloudSave(equipment)
				}
				warningDialog.Show()
			} else {
				doCloudSave(equipment)
			}
		})
		bottomButtons.Add(cloudSaveBtn)

		// перенос проекта в команду для владельца
		proj := appData.Projects[projIdx]
		if proj.CloudID > 0 && proj.OwnerID == currentAuth.UserID {
			moveBtn := widget.NewButtonWithIcon("Перенести в команду", theme.MoveUpIcon(), func() {
				// загрузка команд пользователя
				teams, err := loadTeams(currentAuth.GetToken())
				if err != nil || len(teams) == 0 {
					dialog.ShowInformation("Нет команд", "Вы не состоите ни в одной команде.", w)
					return
				}

				teamNames := []string{"(Личный проект)"}
				teamMap := make(map[string]*uint)
				teamMap["(Личный проект)"] = nil

				for _, t := range teams {
					teamNames = append(teamNames, t.Name)
					tid := t.ID
					teamMap[t.Name] = &tid
				}

				selectTeam := widget.NewSelect(teamNames, nil)
				if proj.TeamID != nil {
					for _, t := range teams {
						if t.ID == *proj.TeamID {
							selectTeam.SetSelected(t.Name)
							break
						}
					}
				} else {
					selectTeam.SetSelected("(Личный проект)")
				}

				dialog.ShowForm("Перенос проекта", "Перенести", "Отмена",
					[]*widget.FormItem{widget.NewFormItem("Команда:", selectTeam)},
					func(ok bool) {
						if ok && selectTeam.Selected != "" {
							tid := teamMap[selectTeam.Selected]
							err := moveProjectToTeam(proj.CloudID, tid, currentAuth.GetToken())
							if err != nil {
								dialog.ShowError(err, w)
							} else {
								// обновление локальных данных
								appData.Projects[projIdx].TeamID = tid
								_ = saveProjects(appData)
								dialog.ShowInformation("Готово", "Проект перенесён.", w)
								showProject(w, proj.Name) // перерисовка экрана проекта
							}
						}
					}, w)
			})
			bottomButtons.Add(moveBtn)
		}
	}

	content := container.NewBorder(
		container.NewVBox(toolbarTop, toolbarActions, title, widget.NewSeparator()),
		container.NewVBox(footer, bottomButtons),
		nil, nil,
		scrollable,
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
}
