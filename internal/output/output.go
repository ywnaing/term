package output

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var NoColor bool

func Title(text string) string {
	if NoColor {
		return text
	}
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(text)
}

func Error(text string) string {
	if NoColor {
		return text
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(text)
}

func PrintSection(title string) {
	fmt.Println(Title(title))
	fmt.Println()
}
