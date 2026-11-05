package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const ColorNameCardBackground fyne.ThemeColorName = "cardBackground"

type modernTheme struct {
	variant fyne.ThemeVariant
}

var _ fyne.Theme = (*modernTheme)(nil)

func (m *modernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// выбор принудительного или переданного варианта темы
	v := m.variant
	if v != theme.VariantDark && v != theme.VariantLight {
		v = variant
	}

	if v == theme.VariantDark {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 15, G: 23, B: 42, A: 255} // Slate 900
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 51, G: 65, B: 85, A: 255} // Slate 700
		case ColorNameCardBackground:
			return color.NRGBA{R: 30, G: 41, B: 59, A: 255} // Slate 800
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 56, G: 189, B: 248, A: 255} // Sky 400 (Cyan)
		case theme.ColorNameForeground:
			return color.NRGBA{R: 248, G: 250, B: 252, A: 255} // Slate 50
		case theme.ColorNameMenuBackground:
			return color.NRGBA{R: 15, G: 23, B: 42, A: 255}
		case theme.ColorNameSeparator:
			return color.NRGBA{R: 51, G: 65, B: 85, A: 255}
		case theme.ColorNameButton:
			return color.NRGBA{R: 30, G: 41, B: 59, A: 255}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 56, G: 189, B: 248, A: 255}
		case theme.ColorNameError:
			return color.NRGBA{R: 239, G: 68, B: 68, A: 255}
		default:
			return theme.DefaultTheme().Color(name, theme.VariantDark)
		}
	} else {
		// светлая тема
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 248, G: 250, B: 252, A: 255} // Slate 50
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // White
		case ColorNameCardBackground:
			return color.NRGBA{R: 241, G: 245, B: 249, A: 255} // Slate 100
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 2, G: 132, B: 199, A: 255} // Sky 600
		case theme.ColorNameForeground:
			return color.NRGBA{R: 15, G: 23, B: 42, A: 255} // Slate 900
		case theme.ColorNameMenuBackground:
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		case theme.ColorNameSeparator:
			return color.NRGBA{R: 226, G: 232, B: 240, A: 255} // Slate 200
		case theme.ColorNameButton:
			return color.NRGBA{R: 241, G: 245, B: 249, A: 255}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 14, G: 165, B: 233, A: 255} // Sky 500
		case theme.ColorNameError:
			return color.NRGBA{R: 220, G: 38, B: 38, A: 255} // Red 600
		default:
			return theme.DefaultTheme().Color(name, theme.VariantLight)
		}
	}
}

func (m *modernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *modernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *modernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNameHeadingText:
		return 20
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 6
	case theme.SizeNameScrollBar:
		return 10
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func newModernDarkTheme() fyne.Theme {
	return &modernTheme{variant: theme.VariantDark}
}

func newModernLightTheme() fyne.Theme {
	return &modernTheme{variant: theme.VariantLight}
}

