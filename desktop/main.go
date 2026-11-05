package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// константы
const (
	windowWidth  = 1100
	windowHeight = 750
)

var equipmentTypes = []string{
	"Насосы",
	"Конвейер",
	"Вертикальный аппарат",
	"Горизонтальная емкость",
	"Трубчатый теплообменник",
	"Тарельчатая колонна",
	"Коробчатая технологическая печь",
	"Центробежный компрессор",
}

func main() {
	fmt.Println("Запуск десктопного приложения...")

	// восстановление токена авторизации из файла
	loadAuthState()

	myApp := app.New()
	myApp.Settings().SetTheme(newModernDarkTheme())
	myWindow := myApp.NewWindow("ConstructMaterialAI: Учёт оборудования")
	myWindow.Resize(fyne.NewSize(windowWidth, windowHeight))

	showStartScreen(myWindow)

	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()

	fmt.Println("Приложение закрыто")
}
