package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// константы excel

const (
	sheetPumps       = "Насосы"
	sheetConveyor    = "Конвейер"
	sheetVessels     = "Вертикальные аппараты"
	sheetDrums       = "Горизонтальные емкости"
	sheetUTubes      = "Трубчатые теплообменники"
	sheetTower       = "Тарельчатые колонны"
	sheetBoxFurnaces  = "Коробчатые печи"
	sheetCompressors  = "Центробежные компрессоры"
	sheetTitlePage    = "Титульный лист"
	defaultSheetFix   = "Sheet1" // excelize создаёт Sheet1 по умолчанию
)

// порядок листов
var sheetOrder = []string{sheetPumps, sheetConveyor, sheetVessels, sheetDrums, sheetUTubes, sheetTower, sheetBoxFurnaces, sheetCompressors}

// заголовки для каждого листа и признак обязательности колонки
type colDef struct {
	Title    string
	Required bool
}

var pumpColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Расход (м³/ч)", true},
	{"Напор (м)", true},
	{"Частота (об/мин)", false},
	{"Уд. вес", false},
	{"Мощность (кВт)", false},
	{"Вес (кг)", false},
}

var conveyorColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Длина (м)", true},
	{"Ширина (мм)", true},
	{"Производительность (т/ч)", false},
	{"Вес (кг)", false},
}

var vesselColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Диаметр (мм)", true},
	{"Высота T/T (мм)", true},
	{"Давление (МПа)", false},
	{"Темп. (°C)", false},
	{"Высота юбки (мм)", false},
	{"Высота опор (мм)", false},
	{"Вес (кг)", false},
}

var drumColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Диаметр (мм)", true},
	{"Длина T/T (мм)", true},
	{"Давление (МПа)", false},
	{"Темп. (°C)", false},
	{"Вес (кг)", false},
}

var utubeColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Диаметр кожуха (мм)", true},
	{"Диам. труб (мм)", true},
	{"Длина труб (мм)", true},
	{"Давл. в трубах (МПа)", false},
	{"Пл. теплообм. (м2)", false},
	{"Вес (кг)", false},
}

var towerColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Диаметр аппарата (м)", true},
	{"Кол-во тарелок", true},
	{"Длина (T/T, м)", false},
	{"Давление (МПа)", false},
	{"Вес (кг)", false},
}

var boxFurnaceColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Тепловая мощность (МВт)", true},
	{"Расход сырья (л/с)", true},
	{"Давление змеевика (кПа)", false},
	{"Расч. температура (°C)", false},
	{"Вес (кг)", false},
}

var compressorColumns = []colDef{
	{"Тэг", true},
	{"Кол-во", true},
	{"Расход на всасывании", true},
	{"Давл. всасыв. (кПа)", true},
	{"Давл. нагнет. (кПа)", true},
	{"Мощность привода", false},
	{"Вес (кг)", false},
}

func columnsForSheet(sheet string) []colDef {
	switch sheet {
	case sheetPumps:
		return pumpColumns
	case sheetConveyor:
		return conveyorColumns
	case sheetVessels:
		return vesselColumns
	case sheetDrums:
		return drumColumns
	case sheetUTubes:
		return utubeColumns
	case sheetTower:
		return towerColumns
	case sheetBoxFurnaces:
		return boxFurnaceColumns
	case sheetCompressors:
		return compressorColumns
	}
	return nil
}

// ошибка валидации при импорте excel
type ImportError struct {
	Sheet  string
	Row    int
	Column string
	Msg    string
}

func (e ImportError) String() string {
	return fmt.Sprintf("Лист «%s», строка %d: в колонке «%s» %s", e.Sheet, e.Row, e.Column, e.Msg)
}

func createTitlePage(f *excelize.File, equipment []Equipment) error {
	if err := f.SetSheetName(defaultSheetFix, sheetTitlePage); err != nil {
		return fmt.Errorf("ошибка переименования листа: %w", err)
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
	})
	if err != nil {
		return err
	}

	linkStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color:     "#0000FF",
			Underline: "single",
			Size:      14,
		},
	})
	if err != nil {
		return err
	}

	weightStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 14,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})
	if err != nil {
		return err
	}

	_ = f.SetCellValue(sheetTitlePage, "A1", "Единицы оборудования проекта")
	_ = f.SetCellStyle(sheetTitlePage, "A1", "A1", titleStyle)

	_ = f.SetColWidth(sheetTitlePage, "A", "A", 40)
	_ = f.SetColWidth(sheetTitlePage, "B", "B", 20)
	_ = f.SetColWidth(sheetTitlePage, "D", "D", 35)
	_ = f.SetColWidth(sheetTitlePage, "E", "E", 20)

	// подсчет весов по типам
	weightsByType := make(map[string]float64)
	var totalWeight float64
	for _, eq := range equipment {
		w := eq.CalculatedWeight * float64(eq.Quantity)
		totalWeight += w
		var sName string
		switch eq.Type {
		case "Насосы":
			sName = sheetPumps
		case "Конвейер":
			sName = sheetConveyor
		case "Вертикальный аппарат":
			sName = sheetVessels
		case "Горизонтальная емкость":
			sName = sheetDrums
		case "Трубчатый теплообменник":
			sName = sheetUTubes
		case "Тарельчатая колонна":
			sName = sheetTower
		case "Коробчатая технологическая печь":
			sName = sheetBoxFurnaces
		case "Центробежный компрессор":
			sName = sheetCompressors
		}
		if sName != "" {
			weightsByType[sName] += w
		}
	}

	startRow := 3
	for i, sheet := range sheetOrder {
		row := startRow + i
		cellA := fmt.Sprintf("A%d", row)
		cellB := fmt.Sprintf("B%d", row)

		displayName := sheet
		if sheet == sheetConveyor {
			displayName = "Конвейеры"
		}
		_ = f.SetCellValue(sheetTitlePage, cellA, displayName)

		link := fmt.Sprintf("'%s'!A1", sheet)
		_ = f.SetCellHyperLink(sheetTitlePage, cellA, link, "Location")
		_ = f.SetCellStyle(sheetTitlePage, cellA, cellA, linkStyle)

		if len(equipment) > 0 {
			w := weightsByType[sheet]
			_ = f.SetCellValue(sheetTitlePage, cellB, w)
			_ = f.SetCellStyle(sheetTitlePage, cellB, cellB, weightStyle)
		}
	}

	if len(equipment) > 0 {
		cellLabel := "D3"
		cellVal := "E3"

		_ = f.SetCellValue(sheetTitlePage, cellLabel, "Итоговый вес проекта (кг)")
		_ = f.SetCellValue(sheetTitlePage, cellVal, totalWeight)

		_ = f.SetCellStyle(sheetTitlePage, cellLabel, cellLabel, titleStyle)
		_ = f.SetCellStyle(sheetTitlePage, cellVal, cellVal, titleStyle)
	}

	return nil
}

// генерация пустого шаблона xlsx с заголовками
func generateTemplate(filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// стиль для обязательных заголовков
	requiredStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FF0000",
			Size:  11,
		},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания стиля: %w", err)
	}

	// стиль для обычных заголовков
	normalStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания стиля: %w", err)
	}

	if err := createTitlePage(f, nil); err != nil {
		return err
	}

	for _, sheet := range sheetOrder {
		if _, err := f.NewSheet(sheet); err != nil {
			return fmt.Errorf("ошибка создания листа «%s»: %w", sheet, err)
		}

		cols := columnsForSheet(sheet)
		for j, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(j+1, 1)
			if err := f.SetCellValue(sheet, cell, col.Title); err != nil {
				return fmt.Errorf("ошибка записи заголовка: %w", err)
			}
			if col.Required {
				_ = f.SetCellStyle(sheet, cell, cell, requiredStyle)
			} else {
				_ = f.SetCellStyle(sheet, cell, cell, normalStyle)
			}
		}

		for j := range cols {
			colName, _ := excelize.ColumnNumberToName(j + 1)
			_ = f.SetColWidth(sheet, colName, colName, 18)
		}
	}

	return f.SaveAs(filePath)
}

// экспорт оборудования проекта в xlsx
func exportProject(filePath string, equipment []Equipment) error {
	f := excelize.NewFile()
	defer f.Close()

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
	})
	if err != nil {
		return fmt.Errorf("ошибка создания стиля: %w", err)
	}

	if err := createTitlePage(f, equipment); err != nil {
		return err
	}

	// создание листов и запись заголовков
	for _, sheet := range sheetOrder {
		if _, err := f.NewSheet(sheet); err != nil {
			return fmt.Errorf("ошибка создания листа «%s»: %w", sheet, err)
		}

		cols := columnsForSheet(sheet)
		for j, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(j+1, 1)
			_ = f.SetCellValue(sheet, cell, col.Title)
			_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
		}
		for j := range cols {
			colName, _ := excelize.ColumnNumberToName(j + 1)
			_ = f.SetColWidth(sheet, colName, colName, 18)
		}
	}

	// счетчики строк по листам
	rowCounters := map[string]int{
		sheetPumps:       2,
		sheetConveyor:    2,
		sheetVessels:     2,
		sheetDrums:       2,
		sheetUTubes:      2,
		sheetTower:       2,
		sheetBoxFurnaces: 2,
		sheetCompressors: 2,
	}

	for _, eq := range equipment {
		var sheet string
		switch eq.Type {
		case "Насосы":
			sheet = sheetPumps
		case "Конвейер":
			sheet = sheetConveyor
		case "Вертикальный аппарат":
			sheet = sheetVessels
		case "Горизонтальная емкость":
			sheet = sheetDrums
		case "Трубчатый теплообменник":
			sheet = sheetUTubes
		case "Тарельчатая колонна":
			sheet = sheetTower
		case "Коробчатая технологическая печь":
			sheet = sheetBoxFurnaces
		case "Центробежный компрессор":
			sheet = sheetCompressors
		default:
			continue
		}

		row := rowCounters[sheet]
		rowCounters[sheet]++

		// общие параметры: тэг и количество
		setCell(f, sheet, 1, row, eq.Tag)
		setCell(f, sheet, 2, row, eq.Quantity)

		switch eq.Type {
		case "Насосы":
			setCellOptFloat(f, sheet, 3, row, eq.FlowRate)
			setCellOptFloat(f, sheet, 4, row, eq.FluidHead)
			setCellOptFloat(f, sheet, 5, row, eq.RPM)
			setCellOptFloat(f, sheet, 6, row, eq.SpecGravity)
			setCellOptFloat(f, sheet, 7, row, eq.PowerKW)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 8, row, eq.CalculatedWeight)
			}

		case "Конвейер":
			setCellOptFloat(f, sheet, 3, row, eq.ConveyorLength)
			setCellOptFloat(f, sheet, 4, row, eq.BeltWidth)
			setCellOptFloat(f, sheet, 5, row, eq.ConveyorFlowRate)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 6, row, eq.CalculatedWeight)
			}

		case "Вертикальный аппарат":
			setCellOptFloat(f, sheet, 3, row, eq.VesselDiameter)
			setCellOptFloat(f, sheet, 4, row, eq.VesselTangentToTangentHeight)
			setCellOptFloat(f, sheet, 5, row, eq.DesignGaugePressure)
			setCellOptFloat(f, sheet, 6, row, eq.DesignTemperature)
			setCellOptFloat(f, sheet, 7, row, eq.SkirtHeight)
			setCellOptFloat(f, sheet, 8, row, eq.VesselLegHeight)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 9, row, eq.CalculatedWeight)
			}

		case "Горизонтальная емкость":
			setCellOptFloat(f, sheet, 3, row, eq.VesselDiameter)
			setCellOptFloat(f, sheet, 4, row, eq.DesignTangentToTangentLength)
			setCellOptFloat(f, sheet, 5, row, eq.DesignGaugePressure)
			setCellOptFloat(f, sheet, 6, row, eq.DesignTemperature)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 7, row, eq.CalculatedWeight)
			}

		case "Трубчатый теплообменник":
			setCellOptFloat(f, sheet, 3, row, eq.ShellDiameter)
			setCellOptFloat(f, sheet, 4, row, eq.TubeOutDiameter)
			setCellOptFloat(f, sheet, 5, row, eq.TubeLen)
			setCellOptFloat(f, sheet, 6, row, eq.TubeDesPres)
			setCellOptFloat(f, sheet, 7, row, eq.HeatArea)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 8, row, eq.CalculatedWeight)
			}

		case "Тарельчатая колонна":
			setCellOptFloat(f, sheet, 3, row, eq.VesselDiameter)
			setCellOptFloat(f, sheet, 4, row, eq.NumberOfTrays)
			setCellOptFloat(f, sheet, 5, row, eq.DesignTangentToTangentLength)
			setCellOptFloat(f, sheet, 6, row, eq.DesignGaugePressure)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 7, row, eq.CalculatedWeight)
			}

		case "Коробчатая технологическая печь":
			setCellOptFloat(f, sheet, 3, row, eq.Duty)
			setCellOptFloat(f, sheet, 4, row, eq.StandardGasFlowRate)
			setCellOptFloat(f, sheet, 5, row, eq.DesignGaugePressure)
			setCellOptFloat(f, sheet, 6, row, eq.DesignTemperature)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 7, row, eq.CalculatedWeight)
			}

		case "Центробежный компрессор":
			setCellOptFloat(f, sheet, 3, row, eq.ActualGasFlowRateInlet)
			setCellOptFloat(f, sheet, 4, row, eq.DesignGaugePressureInlet)
			setCellOptFloat(f, sheet, 5, row, eq.DesignGaugePressureOutlet)
			setCellOptFloat(f, sheet, 6, row, eq.DriverPower)
			if eq.CalculatedWeight > 0 {
				setCell(f, sheet, 7, row, eq.CalculatedWeight)
			}
		}
	}

	return f.SaveAs(filePath)
}

// чтение xlsx и возврат оборудования с ошибками валидации
func importProject(filePath string) ([]Equipment, []ImportError) {
	var equipment []Equipment
	var errors []ImportError

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		errors = append(errors, ImportError{
			Sheet:  "—",
			Row:    0,
			Column: "—",
			Msg:    fmt.Sprintf("невозможно открыть файл: %v", err),
		})
		return nil, errors
	}
	defer f.Close()

	// обработка каждого листа
	for _, sheet := range sheetOrder {
		rows, err := f.GetRows(sheet)
		if err != nil {
			// лист может отсутствовать
			continue
		}

		cols := columnsForSheet(sheet)

		// пропуск строки заголовков
		for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
			excelRow := rowIdx + 1
			cells := rows[rowIdx]

			// пропуск пустых строк
			if isEmptyRow(cells) {
				continue
			}

			eq := Equipment{}
			hasError := false

			// тэг
			tag := getCellStr(cells, 0)
			if tag == "" {
				errors = append(errors, ImportError{
					Sheet:  sheet,
					Row:    excelRow,
					Column: cols[0].Title,
					Msg:    "не заполнен обязательный параметр «Тэг»",
				})
				hasError = true
			}
			eq.Tag = tag

			// количество
			qtyStr := getCellStr(cells, 1)
			qty, qtyErr := strconv.Atoi(strings.TrimSpace(qtyStr))
			if qtyErr != nil || qty < 1 {
				if qtyStr == "" {
					errors = append(errors, ImportError{
						Sheet:  sheet,
						Row:    excelRow,
						Column: cols[1].Title,
						Msg:    "не заполнен обязательный параметр «Кол-во»",
					})
				} else {
					errors = append(errors, ImportError{
						Sheet:  sheet,
						Row:    excelRow,
						Column: cols[1].Title,
						Msg:    fmt.Sprintf("ожидалось целое число ≥ 1, найдено «%s»", qtyStr),
					})
				}
				hasError = true
			} else {
				eq.Quantity = qty
			}

			switch sheet {
			case sheetPumps:
				eq.Type = "Насосы"
				// расход
				eq.FlowRate, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				// напор
				eq.FluidHead, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				// частота
				eq.RPM, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, false, &errors, hasError)
				// удельный вес
				eq.SpecGravity, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				// мощность
				eq.PowerKW, hasError = parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				// вес
				weight, _ := parseImportFloat(cells, 7, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

			case sheetConveyor:
				eq.Type = "Конвейер"
				// длина
				eq.ConveyorLength, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				// ширина
				eq.BeltWidth, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				// производительность
				eq.ConveyorFlowRate, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, false, &errors, hasError)
				// вес
				weight, _ := parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

			case sheetVessels:
				eq.Type = "Вертикальный аппарат"
				// диаметр
				eq.VesselDiameter, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				// высота
				eq.VesselTangentToTangentHeight, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				// давление
				eq.DesignGaugePressure, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, false, &errors, hasError)
				// температура
				eq.DesignTemperature, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				// высота юбки
				eq.SkirtHeight, hasError = parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				// высота опор
				eq.VesselLegHeight, hasError = parseImportFloat(cells, 7, cols, sheet, excelRow, false, &errors, hasError)
				// вес
				weight, _ := parseImportFloat(cells, 8, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

				if eq.SkirtHeight != nil && eq.VesselLegHeight != nil {
					errors = append(errors, ImportError{
						Sheet:  sheet,
						Row:    excelRow,
						Column: "Высота юбки / Высота опор",
						Msg:    "нельзя указывать одновременно высоту юбки и высоту опор",
					})
					hasError = true
				}

			case sheetDrums:
				eq.Type = "Горизонтальная емкость"
				// диаметр
				eq.VesselDiameter, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				// длина
				eq.DesignTangentToTangentLength, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				// давление
				eq.DesignGaugePressure, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, false, &errors, hasError)
				// температура
				eq.DesignTemperature, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				// вес
				weight, _ := parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

			case sheetUTubes:
				eq.Type = "Трубчатый теплообменник"
				// диаметр кожуха
				eq.ShellDiameter, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				// диаметр труб
				eq.TubeOutDiameter, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				// длина труб
				eq.TubeLen, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, true, &errors, hasError)
				// давление в трубах
				eq.TubeDesPres, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				// площадь теплообмена
				eq.HeatArea, hasError = parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				// вес
				weight, _ := parseImportFloat(cells, 7, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

			case sheetTower:
				eq.Type = "Тарельчатая колонна"
				eq.VesselDiameter, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				eq.NumberOfTrays, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				eq.DesignTangentToTangentLength, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, false, &errors, hasError)
				eq.DesignGaugePressure, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				weight, _ := parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

			case sheetBoxFurnaces:
				eq.Type = "Коробчатая технологическая печь"
				eq.Duty, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				eq.StandardGasFlowRate, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				eq.DesignGaugePressure, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, false, &errors, hasError)
				eq.DesignTemperature, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				weight, _ := parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}

			case sheetCompressors:
				eq.Type = "Центробежный компрессор"
				// расход
				eq.ActualGasFlowRateInlet, hasError = parseImportFloat(cells, 2, cols, sheet, excelRow, true, &errors, hasError)
				// давление всасывания
				eq.DesignGaugePressureInlet, hasError = parseImportFloat(cells, 3, cols, sheet, excelRow, true, &errors, hasError)
				// давление нагнетания
				eq.DesignGaugePressureOutlet, hasError = parseImportFloat(cells, 4, cols, sheet, excelRow, true, &errors, hasError)
				// мощность привода
				eq.DriverPower, hasError = parseImportFloat(cells, 5, cols, sheet, excelRow, false, &errors, hasError)
				// вес
				weight, _ := parseImportFloat(cells, 6, cols, sheet, excelRow, false, &errors, hasError)
				if weight != nil {
					eq.CalculatedWeight = *weight
				}
			}

			if !hasError {
				equipment = append(equipment, eq)
			}
		}
	}

	return equipment, errors
}

func setCell(f *excelize.File, sheet string, col, row int, value interface{}) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, value)
}

func setCellOptFloat(f *excelize.File, sheet string, col, row int, val *float64) {
	if val != nil {
		setCell(f, sheet, col, row, *val)
	}
}

func getCellStr(cells []string, idx int) string {
	if idx >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[idx])
}

func isEmptyRow(cells []string) bool {
	for _, c := range cells {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// парсинг вещественного числа из ячейки с валидацией
func parseImportFloat(
	cells []string,
	colIdx int,
	cols []colDef,
	sheet string,
	excelRow int,
	required bool,
	errors *[]ImportError,
	currentHasError bool,
) (*float64, bool) {
	raw := getCellStr(cells, colIdx)
	colTitle := "?"
	if colIdx < len(cols) {
		colTitle = cols[colIdx].Title
	}

	if raw == "" {
		if required {
			*errors = append(*errors, ImportError{
				Sheet:  sheet,
				Row:    excelRow,
				Column: colTitle,
				Msg:    fmt.Sprintf("не заполнен обязательный параметр «%s»", colTitle),
			})
			return nil, true
		}
		return nil, currentHasError
	}

	// замена запятой на точку
	raw = strings.ReplaceAll(raw, ",", ".")
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		*errors = append(*errors, ImportError{
			Sheet:  sheet,
			Row:    excelRow,
			Column: colTitle,
			Msg:    fmt.Sprintf("ожидалось число, найдено «%s»", getCellStr(cells, colIdx)),
		})
		return nil, true
	}

	return &val, currentHasError
}
