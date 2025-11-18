package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

func (m CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// 强制使用浅色主题，忽略系统设置
	switch name {
	case theme.ColorNameBackground:
		return color.White // 强制白色背景
	case theme.ColorNameForeground:
		return color.Black // 黑色前景文字
	case theme.ColorNameDisabled:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 128, G: 128, B: 128, A: 255}
	case theme.ColorNameButton:
		return color.RGBA{R: 240, G: 240, B: 240, A: 255} // 浅灰按钮背景
	case theme.ColorNameInputBackground:
		return color.White // 输入框白色背景
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 250, G: 250, B: 250, A: 255}
	default:
		// 强制使用浅色变体
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (m CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	default:
		return theme.DefaultTheme().Size(name)
	}
}

