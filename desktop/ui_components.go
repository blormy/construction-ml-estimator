package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// элементы ui
type projectCard struct {
	project   Project
	container *fyne.Container
	bg        *canvas.Rectangle
	accent    *canvas.Rectangle
}

func (c *projectCard) refreshTheme() {
	v := fyne.CurrentApp().Settings().ThemeVariant()
	if c.bg != nil {
		c.bg.FillColor = theme.Current().Color(ColorNameCardBackground, v)
		c.bg.Refresh()
	}
	if c.accent != nil {
		c.accent.FillColor = theme.PrimaryColor()
		c.accent.Refresh()
	}
}

type equipmentRow struct {
	typeSelect *widget.Select
	tagEntry   *widget.Entry
	qtyEntry   *widget.Entry

	flowLabel        *canvas.Text
	flowEntry        *widget.Entry
	headLabel        *canvas.Text
	headEntry        *widget.Entry
	rpmLabel         *canvas.Text
	rpmEntry         *widget.Entry
	specGravityLabel *canvas.Text
	specGravityEntry *widget.Entry
	powerLabel       *canvas.Text
	powerEntry       *widget.Entry

	conveyorLengthLabel   *canvas.Text
	conveyorLengthEntry   *widget.Entry
	beltWidthLabel        *canvas.Text
	beltWidthEntry        *widget.Entry
	conveyorFlowRateLabel *canvas.Text
	conveyorFlowRateEntry *widget.Entry

	vesselDiameterLabel               *canvas.Text
	vesselDiameterEntry               *widget.Entry
	designTangentToTangentLengthLabel *canvas.Text
	designTangentToTangentLengthEntry *widget.Entry
	vesselTangentToTangentHeightLabel *canvas.Text
	vesselTangentToTangentHeightEntry *widget.Entry
	designGaugePressureLabel          *canvas.Text
	designGaugePressureEntry          *widget.Entry
	designTemperatureLabel            *canvas.Text
	designTemperatureEntry            *widget.Entry
	skirtHeightLabel                  *canvas.Text
	skirtHeightEntry                  *widget.Entry
	vesselLegHeightLabel              *canvas.Text
	vesselLegHeightEntry              *widget.Entry
	numberOfTraysLabel                *canvas.Text
	numberOfTraysEntry                *widget.Entry

	shellDiameterLabel   *canvas.Text
	shellDiameterEntry   *widget.Entry
	tubeOutDiameterLabel *canvas.Text
	tubeOutDiameterEntry *widget.Entry
	tubeLenLabel         *canvas.Text
	tubeLenEntry         *widget.Entry
	tubeDesPresLabel     *canvas.Text
	tubeDesPresEntry     *widget.Entry
	heatAreaLabel        *canvas.Text
	heatAreaEntry        *widget.Entry

	dutyLabel                *canvas.Text
	dutyEntry                *widget.Entry
	standardGasFlowRateLabel *canvas.Text
	standardGasFlowRateEntry *widget.Entry

	actualGasFlowRateInletLabel    *canvas.Text
	actualGasFlowRateInletEntry    *widget.Entry
	designGaugePressureInletLabel  *canvas.Text
	designGaugePressureInletEntry  *widget.Entry
	designGaugePressureOutletLabel *canvas.Text
	designGaugePressureOutletEntry *widget.Entry
	driverPowerLabel               *canvas.Text
	driverPowerEntry               *widget.Entry

	fieldsContainer *fyne.Container
	resultLabel     *widget.Label
	expandBtn       *widget.Button
	deleteBtn       *widget.Button
	container       *fyne.Container

	cardBg    *canvas.Rectangle
	accentBar *canvas.Rectangle
	expandBg  *canvas.Rectangle
	deleteBg  *canvas.Rectangle
	OnChanged func()
	loading   bool
}

func (r *equipmentRow) dataChanged(isParameter bool) {
	if r.OnChanged == nil || r.loading {
		return
	}
	if isParameter {
		if !strings.HasPrefix(r.resultLabel.Text, "⏳") && r.resultLabel.Text != "—" {
			r.resultLabel.SetText("—")
		}
	}
	r.OnChanged()
}


func (r *equipmentRow) refreshTheme() {
	v := fyne.CurrentApp().Settings().ThemeVariant()
	if r.cardBg != nil {
		r.cardBg.FillColor = theme.Current().Color(ColorNameCardBackground, v)
		r.cardBg.Refresh()
	}
	if r.accentBar != nil {
		r.accentBar.FillColor = theme.PrimaryColor()
		r.accentBar.Refresh()
	}
	if r.expandBg != nil {
		r.expandBg.FillColor = theme.Current().Color(theme.ColorNameInputBackground, v)
		r.expandBg.Refresh()
	}
	if r.deleteBg != nil {
		r.deleteBg.FillColor = theme.Current().Color(theme.ColorNameInputBackground, v)
		r.deleteBg.Refresh()
	}
	if r.expandBtn != nil {
		r.expandBtn.Refresh()
	}
	if r.deleteBtn != nil {
		r.deleteBtn.Refresh()
	}

	labels := []*canvas.Text{
		r.flowLabel, r.headLabel, r.rpmLabel, r.specGravityLabel, r.powerLabel,
		r.conveyorLengthLabel, r.beltWidthLabel, r.conveyorFlowRateLabel,
		r.vesselDiameterLabel, r.vesselTangentToTangentHeightLabel,
		r.designGaugePressureLabel, r.designTemperatureLabel,
		r.skirtHeightLabel, r.vesselLegHeightLabel,
		r.designTangentToTangentLengthLabel, r.numberOfTraysLabel,
		r.shellDiameterLabel, r.tubeOutDiameterLabel, r.tubeLenLabel,
		r.tubeDesPresLabel, r.heatAreaLabel,
		r.dutyLabel, r.standardGasFlowRateLabel,
		r.actualGasFlowRateInletLabel, r.designGaugePressureInletLabel,
		r.designGaugePressureOutletLabel, r.driverPowerLabel,
	}
	for _, l := range labels {
		if l != nil {
			l.Color = theme.ForegroundColor()
			l.Refresh()
		}
	}

	r.container.Refresh()
}

func (r *equipmentRow) markFieldInvalid(e *widget.Entry, label *canvas.Text, hasError bool) {
	if e == nil {
		return
	}
	if hasError {
		e.SetValidationError(fmt.Errorf("invalid"))
	} else {
		e.SetValidationError(nil)
	}
	setLabelError(label, hasError)
}

func (r *equipmentRow) collectEquipment() (Equipment, error) {
	r.clearValidation()

	eq := Equipment{
		Type: strings.TrimSpace(r.typeSelect.Selected),
		Tag:  strings.TrimSpace(r.tagEntry.Text),
	}

	if eq.Tag == "" {
		r.markFieldInvalid(r.tagEntry, nil, true)
		return eq, fmt.Errorf("Тэг обязателен")
	}

	q, err := strconv.Atoi(strings.TrimSpace(r.qtyEntry.Text))
	if err != nil || q < 1 {
		r.markFieldInvalid(r.qtyEntry, nil, true)
		return eq, fmt.Errorf("Кол-во должно быть >= 1")
	}
	eq.Quantity = q

	switch eq.Type {
	case "Насосы":
		flow, err := parseOptionalFloat(r.flowEntry.Text)
		if err != nil || flow == nil {
			r.markFieldInvalid(r.flowEntry, r.flowLabel, true)
			return eq, fmt.Errorf("Расход обязателен")
		}
		eq.FlowRate = flow

		head, err := parseOptionalFloat(r.headEntry.Text)
		if err != nil || head == nil {
			r.markFieldInvalid(r.headEntry, r.headLabel, true)
			return eq, fmt.Errorf("Напор обязателен")
		}
		eq.FluidHead = head

		eq.RPM, err = parseOptionalFloat(r.rpmEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.rpmEntry, r.rpmLabel, true)
		}
		eq.SpecGravity, err = parseOptionalFloat(r.specGravityEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.specGravityEntry, r.specGravityLabel, true)
		}
		eq.PowerKW, err = parseOptionalFloat(r.powerEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.powerEntry, r.powerLabel, true)
		}

	case "Конвейер":
		conveyorLength, err := parseOptionalFloat(r.conveyorLengthEntry.Text)
		if err != nil || conveyorLength == nil {
			r.markFieldInvalid(r.conveyorLengthEntry, r.conveyorLengthLabel, true)
			return eq, fmt.Errorf("Длина конвейера обязательна")
		}
		eq.ConveyorLength = conveyorLength

		beltWidth, err := parseOptionalFloat(r.beltWidthEntry.Text)
		if err != nil || beltWidth == nil {
			r.markFieldInvalid(r.beltWidthEntry, r.beltWidthLabel, true)
			return eq, fmt.Errorf("Ширина ленты обязательна")
		}
		eq.BeltWidth = beltWidth

		eq.ConveyorFlowRate, err = parseOptionalFloat(r.conveyorFlowRateEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.conveyorFlowRateEntry, r.conveyorFlowRateLabel, true)
		}

	case "Вертикальный аппарат":
		vesselDiameter, err := parseOptionalFloat(r.vesselDiameterEntry.Text)
		if err != nil || vesselDiameter == nil {
			r.markFieldInvalid(r.vesselDiameterEntry, r.vesselDiameterLabel, true)
			return eq, fmt.Errorf("Диаметр обязателен")
		}
		eq.VesselDiameter = vesselDiameter

		vesselHeight, err := parseOptionalFloat(r.vesselTangentToTangentHeightEntry.Text)
		if err != nil || vesselHeight == nil {
			r.markFieldInvalid(r.vesselTangentToTangentHeightEntry, r.vesselTangentToTangentHeightLabel, true)
			return eq, fmt.Errorf("Высота обязательна")
		}
		eq.VesselTangentToTangentHeight = vesselHeight

		eq.DesignGaugePressure, err = parseOptionalFloat(r.designGaugePressureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designGaugePressureEntry, r.designGaugePressureLabel, true)
		}
		eq.DesignTemperature, err = parseOptionalFloat(r.designTemperatureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designTemperatureEntry, r.designTemperatureLabel, true)
		}
		eq.SkirtHeight, err = parseOptionalFloat(r.skirtHeightEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.skirtHeightEntry, r.skirtHeightLabel, true)
		}
		eq.VesselLegHeight, err = parseOptionalFloat(r.vesselLegHeightEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.vesselLegHeightEntry, r.vesselLegHeightLabel, true)
		}

		if eq.SkirtHeight != nil && eq.VesselLegHeight != nil {
			r.markFieldInvalid(r.skirtHeightEntry, r.skirtHeightLabel, true)
			r.markFieldInvalid(r.vesselLegHeightEntry, r.vesselLegHeightLabel, true)
			return eq, fmt.Errorf("Нельзя указывать одновременно высоту юбки и опор")
		}

	case "Горизонтальная емкость":
		vesselDiameter, err := parseOptionalFloat(r.vesselDiameterEntry.Text)
		if err != nil || vesselDiameter == nil {
			r.markFieldInvalid(r.vesselDiameterEntry, r.vesselDiameterLabel, true)
			return eq, fmt.Errorf("Диаметр обязателен")
		}
		eq.VesselDiameter = vesselDiameter

		designLength, err := parseOptionalFloat(r.designTangentToTangentLengthEntry.Text)
		if err != nil || designLength == nil {
			r.markFieldInvalid(r.designTangentToTangentLengthEntry, r.designTangentToTangentLengthLabel, true)
			return eq, fmt.Errorf("Длина обязательна")
		}
		eq.DesignTangentToTangentLength = designLength

		eq.DesignGaugePressure, err = parseOptionalFloat(r.designGaugePressureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designGaugePressureEntry, r.designGaugePressureLabel, true)
		}
		eq.DesignTemperature, err = parseOptionalFloat(r.designTemperatureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designTemperatureEntry, r.designTemperatureLabel, true)
		}

	case "Тарельчатая колонна":
		vesselDiameter, err := parseOptionalFloat(r.vesselDiameterEntry.Text)
		if err != nil || vesselDiameter == nil {
			r.markFieldInvalid(r.vesselDiameterEntry, r.vesselDiameterLabel, true)
			return eq, fmt.Errorf("Диаметр обязателен")
		}
		eq.VesselDiameter = vesselDiameter

		numberOfTrays, err := parseOptionalFloat(r.numberOfTraysEntry.Text)
		if err != nil || numberOfTrays == nil {
			r.markFieldInvalid(r.numberOfTraysEntry, r.numberOfTraysLabel, true)
			return eq, fmt.Errorf("Количество тарелок обязательно")
		}
		eq.NumberOfTrays = numberOfTrays

		eq.DesignTangentToTangentLength, err = parseOptionalFloat(r.designTangentToTangentLengthEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designTangentToTangentLengthEntry, r.designTangentToTangentLengthLabel, true)
		}
		eq.DesignGaugePressure, err = parseOptionalFloat(r.designGaugePressureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designGaugePressureEntry, r.designGaugePressureLabel, true)
		}

	case "Коробчатая технологическая печь":
		duty, err := parseOptionalFloat(r.dutyEntry.Text)
		if err != nil || duty == nil {
			r.markFieldInvalid(r.dutyEntry, r.dutyLabel, true)
			return eq, fmt.Errorf("Тепловая мощность обязательна")
		}
		eq.Duty = duty

		gasFlow, err := parseOptionalFloat(r.standardGasFlowRateEntry.Text)
		if err != nil || gasFlow == nil {
			r.markFieldInvalid(r.standardGasFlowRateEntry, r.standardGasFlowRateLabel, true)
			return eq, fmt.Errorf("Расход сырья обязателен")
		}
		eq.StandardGasFlowRate = gasFlow

		eq.DesignGaugePressure, err = parseOptionalFloat(r.designGaugePressureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designGaugePressureEntry, r.designGaugePressureLabel, true)
		}
		eq.DesignTemperature, err = parseOptionalFloat(r.designTemperatureEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.designTemperatureEntry, r.designTemperatureLabel, true)
		}

	case "Центробежный компрессор":
		gasFlow, err := parseOptionalFloat(r.actualGasFlowRateInletEntry.Text)
		if err != nil || gasFlow == nil {
			r.markFieldInvalid(r.actualGasFlowRateInletEntry, r.actualGasFlowRateInletLabel, true)
			return eq, fmt.Errorf("Расход на всасывании обязателен")
		}
		eq.ActualGasFlowRateInlet = gasFlow

		pInlet, err := parseOptionalFloat(r.designGaugePressureInletEntry.Text)
		if err != nil || pInlet == nil {
			r.markFieldInvalid(r.designGaugePressureInletEntry, r.designGaugePressureInletLabel, true)
			return eq, fmt.Errorf("Давление на всасывании обязательно")
		}
		eq.DesignGaugePressureInlet = pInlet

		pOutlet, err := parseOptionalFloat(r.designGaugePressureOutletEntry.Text)
		if err != nil || pOutlet == nil {
			r.markFieldInvalid(r.designGaugePressureOutletEntry, r.designGaugePressureOutletLabel, true)
			return eq, fmt.Errorf("Давление нагнетания обязательно")
		}
		eq.DesignGaugePressureOutlet = pOutlet

		eq.DriverPower, err = parseOptionalFloat(r.driverPowerEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.driverPowerEntry, r.driverPowerLabel, true)
		}

	case "Трубчатый теплообменник":
		shellDiameter, err := parseOptionalFloat(r.shellDiameterEntry.Text)
		if err != nil || shellDiameter == nil {
			r.markFieldInvalid(r.shellDiameterEntry, r.shellDiameterLabel, true)
			return eq, fmt.Errorf("Диаметр кожуха обязателен")
		}
		eq.ShellDiameter = shellDiameter

		tubeOutDiameter, err := parseOptionalFloat(r.tubeOutDiameterEntry.Text)
		if err != nil || tubeOutDiameter == nil {
			r.markFieldInvalid(r.tubeOutDiameterEntry, r.tubeOutDiameterLabel, true)
			return eq, fmt.Errorf("Диаметр труб обязателен")
		}
		eq.TubeOutDiameter = tubeOutDiameter

		tubeLen, err := parseOptionalFloat(r.tubeLenEntry.Text)
		if err != nil || tubeLen == nil {
			r.markFieldInvalid(r.tubeLenEntry, r.tubeLenLabel, true)
			return eq, fmt.Errorf("Длина труб обязательна")
		}
		eq.TubeLen = tubeLen

		eq.TubeDesPres, err = parseOptionalFloat(r.tubeDesPresEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.tubeDesPresEntry, r.tubeDesPresLabel, true)
		}
		eq.HeatArea, err = parseOptionalFloat(r.heatAreaEntry.Text)
		if err != nil {
			r.markFieldInvalid(r.heatAreaEntry, r.heatAreaLabel, true)
		}
	}

	return eq, nil
}

// кнопка с поддержкой событий наведения
type HoverButton struct {
	widget.Button
	OnHover func(bool)
}

func (b *HoverButton) MouseIn(*desktop.MouseEvent) {
	if b.OnHover != nil {
		b.OnHover(true)
	}
}

func (b *HoverButton) MouseOut() {
	if b.OnHover != nil {
		b.OnHover(false)
	}
}

func NewHoverButton(label string, icon fyne.Resource, tapped func(), hover func(bool)) *HoverButton {
	b := &HoverButton{OnHover: hover}
	b.Text = label
	b.Icon = icon
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

// создание кнопки с подложкой и эффектом наведения
func NewThemedHoverButton(label string, icon fyne.Resource, tapped func()) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.PrimaryColor())
	bg.CornerRadius = 10

	btn := NewHoverButton(label, icon, tapped, nil)
	btn.Importance = widget.LowImportance

	btn.OnHover = func(hover bool) {
		if hover {
			bg.FillColor = color.NRGBA{R: 2, G: 132, B: 199, A: 255}
		} else {
			bg.FillColor = theme.PrimaryColor()
		}
		bg.Refresh()
	}
	return container.NewStack(bg, btn)
}
func createProjectCard(w fyne.Window, proj Project, onOpen func(), onDelete func()) *projectCard {
	card := &projectCard{project: proj}

	title := widget.NewLabelWithStyle(proj.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Truncation = fyne.TextTruncateEllipsis

	eqCount := proj.EquipmentCount()
	weight := proj.TotalWeight()

	info := widget.NewLabel(fmt.Sprintf("Оборудование: %d | Вес: %.2f кг", eqCount, weight))
	info.TextStyle = fyne.TextStyle{Italic: true}

	var teamBadge *widget.Label
	if proj.TeamID != nil && *proj.TeamID > 0 {
		teamBadge = widget.NewLabelWithStyle("КОМАНДНЫЙ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}

	openBtn := widget.NewButtonWithIcon("Открыть", theme.FolderOpenIcon(), onOpen)
	openBtn.Importance = widget.HighImportance

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), onDelete)
	deleteBtn.Importance = widget.LowImportance

	// блокировка удаления для не-владельца
	if proj.OwnerID > 0 && currentAuth.IsLoggedIn() && proj.OwnerID != currentAuth.UserID {
		deleteBtn.Disable()
	}

	v := fyne.CurrentApp().Settings().ThemeVariant()
	card.bg = canvas.NewRectangle(theme.Current().Color(ColorNameCardBackground, v))
	card.bg.CornerRadius = 12

	card.accent = canvas.NewRectangle(theme.PrimaryColor())
	card.accent.SetMinSize(fyne.NewSize(4, 0))

	vbox := container.NewVBox(title)
	if teamBadge != nil {
		vbox.Add(teamBadge)
	}
	vbox.Add(info)

	content := container.NewPadded(container.NewHBox(
		card.accent,
		vbox,
		layout.NewSpacer(),
		container.NewHBox(openBtn, deleteBtn),
	))

	card.container = container.NewStack(card.bg, content)
	return card
}
func createLabel(text string, required bool) (*canvas.Text, fyne.CanvasObject) {
	t := canvas.NewText(text, theme.ForegroundColor())
	t.TextSize = theme.TextSize() + 1
	t.TextStyle.Bold = true

	var finalObj fyne.CanvasObject
	if !required {
		finalObj = t
	} else {
		ast := canvas.NewText(" *", color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		ast.TextSize = t.TextSize
		ast.TextStyle.Bold = true
		finalObj = container.NewHBox(t, ast)
	}

	// контейнер с фиксированной шириной для выравнивания
	return t, container.NewGridWrap(fyne.NewSize(260, 40), finalObj)
}

func setLabelError(t *canvas.Text, hasError bool) {
	if t == nil {
		return
	}
	fyne.Do(func() {
		t.Color = theme.ForegroundColor()
		if hasError {
			t.TextStyle.Bold = true
		} else {
			// сохранение жирного начертания для предотвращения сдвига текста
			t.TextStyle.Bold = true
		}
		t.Refresh()
	})
}

func (r *equipmentRow) clearValidation() {
	// сброс валидации для всех полей
	entries := []*widget.Entry{
		r.tagEntry, r.qtyEntry, r.flowEntry, r.headEntry, r.rpmEntry,
		r.specGravityEntry, r.powerEntry, r.conveyorLengthEntry, r.beltWidthEntry,
		r.conveyorFlowRateEntry,
		r.vesselDiameterEntry, r.designTangentToTangentLengthEntry,
		r.vesselTangentToTangentHeightEntry, r.designGaugePressureEntry,
		r.designTemperatureEntry, r.skirtHeightEntry, r.vesselLegHeightEntry,
		r.numberOfTraysEntry,
		r.shellDiameterEntry, r.tubeOutDiameterEntry, r.tubeLenEntry,
		r.tubeDesPresEntry, r.heatAreaEntry,
		r.dutyEntry, r.standardGasFlowRateEntry,
		r.actualGasFlowRateInletEntry, r.designGaugePressureInletEntry,
		r.designGaugePressureOutletEntry, r.driverPowerEntry,
	}
	for _, e := range entries {
		if e != nil {
			e.SetValidationError(nil)
		}
	}

	setLabelError(r.flowLabel, false)
	setLabelError(r.headLabel, false)
	setLabelError(r.rpmLabel, false)
	setLabelError(r.specGravityLabel, false)
	setLabelError(r.powerLabel, false)
	setLabelError(r.conveyorLengthLabel, false)
	setLabelError(r.beltWidthLabel, false)
	setLabelError(r.conveyorFlowRateLabel, false)
	setLabelError(r.vesselDiameterLabel, false)
	setLabelError(r.designTangentToTangentLengthLabel, false)
	setLabelError(r.vesselTangentToTangentHeightLabel, false)
	setLabelError(r.designGaugePressureLabel, false)
	setLabelError(r.designTemperatureLabel, false)
	setLabelError(r.skirtHeightLabel, false)
	setLabelError(r.vesselLegHeightLabel, false)
	setLabelError(r.numberOfTraysLabel, false)
	setLabelError(r.shellDiameterLabel, false)
	setLabelError(r.tubeOutDiameterLabel, false)
	setLabelError(r.tubeLenLabel, false)
	setLabelError(r.tubeDesPresLabel, false)
	setLabelError(r.heatAreaLabel, false)
	setLabelError(r.dutyLabel, false)
	setLabelError(r.standardGasFlowRateLabel, false)
	setLabelError(r.actualGasFlowRateInletLabel, false)
	setLabelError(r.designGaugePressureInletLabel, false)
	setLabelError(r.designGaugePressureOutletLabel, false)
	setLabelError(r.driverPowerLabel, false)
}
func buildPumpFields(row *equipmentRow) *fyne.Container {
	flowLabel, flowObj := createLabel("Расход (м³/ч):", true)
	row.flowLabel = flowLabel
	row.flowEntry = widget.NewEntry()
	row.flowEntry.SetPlaceHolder("Введите расход (обязательно)")
	row.flowEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.flowEntry, row.flowLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	headLabel, headObj := createLabel("Напор (м):", true)
	row.headLabel = headLabel
	row.headEntry = widget.NewEntry()
	row.headEntry.SetPlaceHolder("Введите напор (обязательно)")
	row.headEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.headEntry, row.headLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	rpmLabel, rpmObj := createLabel("Частота вращения (об/мин):", false)
	row.rpmLabel = rpmLabel
	row.rpmEntry = widget.NewEntry()
	row.rpmEntry.SetPlaceHolder("Частота вращения (об/мин)")
	row.rpmEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.rpmEntry, row.rpmLabel, err != nil)
		row.dataChanged(true)
	}

	specGravityLabel, specGravityObj := createLabel("Удельный вес:", false)
	row.specGravityLabel = specGravityLabel
	row.specGravityEntry = widget.NewEntry()
	row.specGravityEntry.SetPlaceHolder("Удельный вес")
	row.specGravityEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.specGravityEntry, row.specGravityLabel, err != nil)
		row.dataChanged(true)
	}

	powerLabel, powerObj := createLabel("Мощность (кВт):", false)
	row.powerLabel = powerLabel
	row.powerEntry = widget.NewEntry()
	row.powerEntry.SetPlaceHolder("Мощность (кВт)")
	row.powerEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.powerEntry, row.powerLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		flowObj, row.flowEntry,
		headObj, row.headEntry,
		rpmObj, row.rpmEntry,
		specGravityObj, row.specGravityEntry,
		powerObj, row.powerEntry,
	)
}

func buildConveyorFields(row *equipmentRow) *fyne.Container {
	conveyorLengthLabel, conveyorLengthObj := createLabel("Длина конвейера (м):", true)
	row.conveyorLengthLabel = conveyorLengthLabel
	row.conveyorLengthEntry = widget.NewEntry()
	row.conveyorLengthEntry.SetPlaceHolder("Введите длину (обязательно)")
	row.conveyorLengthEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.conveyorLengthEntry, row.conveyorLengthLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	beltWidthLabel, beltWidthObj := createLabel("Ширина ленты (мм):", true)
	row.beltWidthLabel = beltWidthLabel
	row.beltWidthEntry = widget.NewEntry()
	row.beltWidthEntry.SetPlaceHolder("Введите ширину (обязательно)")
	row.beltWidthEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.beltWidthEntry, row.beltWidthLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	conveyorFlowRateLabel, conveyorFlowRateObj := createLabel("Производительность (т/ч):", false)
	row.conveyorFlowRateLabel = conveyorFlowRateLabel
	row.conveyorFlowRateEntry = widget.NewEntry()
	row.conveyorFlowRateEntry.SetPlaceHolder("Введите производительность")
	row.conveyorFlowRateEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.conveyorFlowRateEntry, row.conveyorFlowRateLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		conveyorLengthObj, row.conveyorLengthEntry,
		beltWidthObj, row.beltWidthEntry,
		conveyorFlowRateObj, row.conveyorFlowRateEntry,
	)
}

func buildVesselFields(row *equipmentRow) *fyne.Container {
	vesselDiameterLabel, vesselDiameterObj := createLabel("Диаметр аппарата (м):", true)
	row.vesselDiameterLabel = vesselDiameterLabel
	row.vesselDiameterEntry = widget.NewEntry()
	row.vesselDiameterEntry.SetPlaceHolder("Диаметр корпуса (м)")
	row.vesselDiameterEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.vesselDiameterEntry, row.vesselDiameterLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	vesselTangentToTangentHeightLabel, vesselTangentToTangentHeightObj := createLabel("Высота (T/T, м):", true)
	row.vesselTangentToTangentHeightLabel = vesselTangentToTangentHeightLabel
	row.vesselTangentToTangentHeightEntry = widget.NewEntry()
	row.vesselTangentToTangentHeightEntry.SetPlaceHolder("Высота Straight Side (м)")
	row.vesselTangentToTangentHeightEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.vesselTangentToTangentHeightEntry, row.vesselTangentToTangentHeightLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	designGaugePressureLabel, designGaugePressureObj := createLabel("Давление (МПа):", false)
	row.designGaugePressureLabel = designGaugePressureLabel
	row.designGaugePressureEntry = widget.NewEntry()
	row.designGaugePressureEntry.SetPlaceHolder("Давление")
	row.designGaugePressureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designGaugePressureEntry, row.designGaugePressureLabel, err != nil)
		row.dataChanged(true)
	}

	designTemperatureLabel, designTemperatureObj := createLabel("Температура (°C):", false)
	row.designTemperatureLabel = designTemperatureLabel
	row.designTemperatureEntry = widget.NewEntry()
	row.designTemperatureEntry.SetPlaceHolder("Температура")
	row.designTemperatureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designTemperatureEntry, row.designTemperatureLabel, err != nil)
		row.dataChanged(true)
	}

	skirtHeightLabel, skirtHeightObj := createLabel("Высота юбки (м):", false)
	row.skirtHeightLabel = skirtHeightLabel
	row.skirtHeightEntry = widget.NewEntry()
	row.skirtHeightEntry.SetPlaceHolder("Высота юбки (м)")
	row.skirtHeightEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.skirtHeightEntry, row.skirtHeightLabel, err != nil)
		row.dataChanged(true)
	}

	vesselLegHeightLabel, vesselLegHeightObj := createLabel("Высота опор (м):", false)
	row.vesselLegHeightLabel = vesselLegHeightLabel
	row.vesselLegHeightEntry = widget.NewEntry()
	row.vesselLegHeightEntry.SetPlaceHolder("Высота опор (м)")
	row.vesselLegHeightEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.vesselLegHeightEntry, row.vesselLegHeightLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		vesselDiameterObj, row.vesselDiameterEntry,
		vesselTangentToTangentHeightObj, row.vesselTangentToTangentHeightEntry,
		designGaugePressureObj, row.designGaugePressureEntry,
		designTemperatureObj, row.designTemperatureEntry,
		skirtHeightObj, row.skirtHeightEntry,
		vesselLegHeightObj, row.vesselLegHeightEntry,
	)
}

func buildDrumFields(row *equipmentRow) *fyne.Container {
	vesselDiameterLabel, vesselDiameterObj := createLabel("Диаметр аппарата (м):", true)
	row.vesselDiameterLabel = vesselDiameterLabel
	row.vesselDiameterEntry = widget.NewEntry()
	row.vesselDiameterEntry.SetPlaceHolder("Введите диаметр (м)")
	row.vesselDiameterEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.vesselDiameterEntry, row.vesselDiameterLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	designTangentToTangentLengthLabel, designTangentToTangentLengthObj := createLabel("Длина (T/T, м):", true)
	row.designTangentToTangentLengthLabel = designTangentToTangentLengthLabel
	row.designTangentToTangentLengthEntry = widget.NewEntry()
	row.designTangentToTangentLengthEntry.SetPlaceHolder("Введите длину (м)")
	row.designTangentToTangentLengthEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designTangentToTangentLengthEntry, row.designTangentToTangentLengthLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	designGaugePressureLabel, designGaugePressureObj := createLabel("Давление (МПа):", false)
	row.designGaugePressureLabel = designGaugePressureLabel
	row.designGaugePressureEntry = widget.NewEntry()
	row.designGaugePressureEntry.SetPlaceHolder("Давление")
	row.designGaugePressureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designGaugePressureEntry, row.designGaugePressureLabel, err != nil)
		row.dataChanged(true)
	}

	designTemperatureLabel, designTemperatureObj := createLabel("Температура (°C):", false)
	row.designTemperatureLabel = designTemperatureLabel
	row.designTemperatureEntry = widget.NewEntry()
	row.designTemperatureEntry.SetPlaceHolder("Температура")
	row.designTemperatureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designTemperatureEntry, row.designTemperatureLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		vesselDiameterObj, row.vesselDiameterEntry,
		designTangentToTangentLengthObj, row.designTangentToTangentLengthEntry,
		designGaugePressureObj, row.designGaugePressureEntry,
		designTemperatureObj, row.designTemperatureEntry,
	)
}

func buildUTubeFields(row *equipmentRow) *fyne.Container {
	shellDiameterLabel, shellDiameterObj := createLabel("Диаметр кожуха (мм):", true)
	row.shellDiameterLabel = shellDiameterLabel
	row.shellDiameterEntry = widget.NewEntry()
	row.shellDiameterEntry.SetPlaceHolder("Введите диаметр (обязательно)")
	row.shellDiameterEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.shellDiameterEntry, row.shellDiameterLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	tubeOutDiameterLabel, tubeOutDiameterObj := createLabel("Диаметр труб (мм):", true)
	row.tubeOutDiameterLabel = tubeOutDiameterLabel
	row.tubeOutDiameterEntry = widget.NewEntry()
	row.tubeOutDiameterEntry.SetPlaceHolder("Введите диаметр (обязательно)")
	row.tubeOutDiameterEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.tubeOutDiameterEntry, row.tubeOutDiameterLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	tubeLenLabel, tubeLenObj := createLabel("Длина труб (мм):", true)
	row.tubeLenLabel = tubeLenLabel
	row.tubeLenEntry = widget.NewEntry()
	row.tubeLenEntry.SetPlaceHolder("Введите длину (обязательно)")
	row.tubeLenEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.tubeLenEntry, row.tubeLenLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	tubeDesPresLabel, tubeDesPresObj := createLabel("Давление в трубах (МПа):", false)
	row.tubeDesPresLabel = tubeDesPresLabel
	row.tubeDesPresEntry = widget.NewEntry()
	row.tubeDesPresEntry.SetPlaceHolder("Давление (МПа)")
	row.tubeDesPresEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.tubeDesPresEntry, row.tubeDesPresLabel, err != nil)
		row.dataChanged(true)
	}

	heatAreaLabel, heatAreaObj := createLabel("Пл. теплообм. (м2):", false)
	row.heatAreaLabel = heatAreaLabel
	row.heatAreaEntry = widget.NewEntry()
	row.heatAreaEntry.SetPlaceHolder("Площадь")
	row.heatAreaEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.heatAreaEntry, row.heatAreaLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		shellDiameterObj, row.shellDiameterEntry,
		tubeOutDiameterObj, row.tubeOutDiameterEntry,
		tubeLenObj, row.tubeLenEntry,
		tubeDesPresObj, row.tubeDesPresEntry,
		heatAreaObj, row.heatAreaEntry,
	)
}

func buildTowerFields(row *equipmentRow) *fyne.Container {
	vesselDiameterLabel, vesselDiameterObj := createLabel("Диаметр аппарата (м):", true)
	row.vesselDiameterLabel = vesselDiameterLabel
	row.vesselDiameterEntry = widget.NewEntry()
	row.vesselDiameterEntry.SetPlaceHolder("Введите диаметр (м)")
	row.vesselDiameterEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.vesselDiameterEntry, row.vesselDiameterLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	numberOfTraysLabel, numberOfTraysObj := createLabel("Количество тарелок:", true)
	row.numberOfTraysLabel = numberOfTraysLabel
	row.numberOfTraysEntry = widget.NewEntry()
	row.numberOfTraysEntry.SetPlaceHolder("Введите количество тарелок")
	row.numberOfTraysEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.numberOfTraysEntry, row.numberOfTraysLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	designTangentToTangentLengthLabel, designTangentToTangentLengthObj := createLabel("Длина (T/T, м):", false)
	row.designTangentToTangentLengthLabel = designTangentToTangentLengthLabel
	row.designTangentToTangentLengthEntry = widget.NewEntry()
	row.designTangentToTangentLengthEntry.SetPlaceHolder("Введите длину (м)")
	row.designTangentToTangentLengthEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designTangentToTangentLengthEntry, row.designTangentToTangentLengthLabel, err != nil)
		row.dataChanged(true)
	}

	designGaugePressureLabel, designGaugePressureObj := createLabel("Давление (МПа):", false)
	row.designGaugePressureLabel = designGaugePressureLabel
	row.designGaugePressureEntry = widget.NewEntry()
	row.designGaugePressureEntry.SetPlaceHolder("Давление")
	row.designGaugePressureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designGaugePressureEntry, row.designGaugePressureLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		vesselDiameterObj, row.vesselDiameterEntry,
		numberOfTraysObj, row.numberOfTraysEntry,
		designTangentToTangentLengthObj, row.designTangentToTangentLengthEntry,
		designGaugePressureObj, row.designGaugePressureEntry,
	)
}

func buildBoxFurnaceFields(row *equipmentRow) *fyne.Container {
	dutyLabel, dutyObj := createLabel("Тепловая мощность (МВт):", true)
	row.dutyLabel = dutyLabel
	row.dutyEntry = widget.NewEntry()
	row.dutyEntry.SetPlaceHolder("Введите мощность (обязательно)")
	row.dutyEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.dutyEntry, row.dutyLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	standardGasFlowRateLabel, standardGasFlowRateObj := createLabel("Расход сырья (л/с):", true)
	row.standardGasFlowRateLabel = standardGasFlowRateLabel
	row.standardGasFlowRateEntry = widget.NewEntry()
	row.standardGasFlowRateEntry.SetPlaceHolder("Введите расход (обязательно)")
	row.standardGasFlowRateEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.standardGasFlowRateEntry, row.standardGasFlowRateLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	designGaugePressureLabel, designGaugePressureObj := createLabel("Давление змеевика (кПа):", false)
	row.designGaugePressureLabel = designGaugePressureLabel
	row.designGaugePressureEntry = widget.NewEntry()
	row.designGaugePressureEntry.SetPlaceHolder("Давление")
	row.designGaugePressureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designGaugePressureEntry, row.designGaugePressureLabel, err != nil)
		row.dataChanged(true)
	}

	designTemperatureLabel, designTemperatureObj := createLabel("Расч. температура (°C):", false)
	row.designTemperatureLabel = designTemperatureLabel
	row.designTemperatureEntry = widget.NewEntry()
	row.designTemperatureEntry.SetPlaceHolder("Температура")
	row.designTemperatureEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designTemperatureEntry, row.designTemperatureLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		dutyObj, row.dutyEntry,
		standardGasFlowRateObj, row.standardGasFlowRateEntry,
		designGaugePressureObj, row.designGaugePressureEntry,
		designTemperatureObj, row.designTemperatureEntry,
	)
}

func buildCentrifugalCompressorFields(row *equipmentRow) *fyne.Container {
	gasFlowLabel, gasFlowObj := createLabel("Расход на всасывании:", true)
	row.actualGasFlowRateInletLabel = gasFlowLabel
	row.actualGasFlowRateInletEntry = widget.NewEntry()
	row.actualGasFlowRateInletEntry.SetPlaceHolder("Введите расход (обязательно)")
	row.actualGasFlowRateInletEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.actualGasFlowRateInletEntry, row.actualGasFlowRateInletLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	pInletLabel, pInletObj := createLabel("Давление на всасывании (кПа):", true)
	row.designGaugePressureInletLabel = pInletLabel
	row.designGaugePressureInletEntry = widget.NewEntry()
	row.designGaugePressureInletEntry.SetPlaceHolder("Введите давление (обязательно)")
	row.designGaugePressureInletEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designGaugePressureInletEntry, row.designGaugePressureInletLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	pOutletLabel, pOutletObj := createLabel("Давление нагнетания (кПа):", true)
	row.designGaugePressureOutletLabel = pOutletLabel
	row.designGaugePressureOutletEntry = widget.NewEntry()
	row.designGaugePressureOutletEntry.SetPlaceHolder("Введите давление (обязательно)")
	row.designGaugePressureOutletEntry.OnChanged = func(s string) {
		val, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.designGaugePressureOutletEntry, row.designGaugePressureOutletLabel, err != nil || val == nil)
		row.dataChanged(true)
	}

	driverPowerLabel, driverPowerObj := createLabel("Мощность привода:", false)
	row.driverPowerLabel = driverPowerLabel
	row.driverPowerEntry = widget.NewEntry()
	row.driverPowerEntry.SetPlaceHolder("Мощность")
	row.driverPowerEntry.OnChanged = func(s string) {
		_, err := parseOptionalFloat(s)
		row.markFieldInvalid(row.driverPowerEntry, row.driverPowerLabel, err != nil)
		row.dataChanged(true)
	}

	return container.New(layout.NewFormLayout(),
		gasFlowObj, row.actualGasFlowRateInletEntry,
		pInletObj, row.designGaugePressureInletEntry,
		pOutletObj, row.designGaugePressureOutletEntry,
		driverPowerObj, row.driverPowerEntry,
	)
}
