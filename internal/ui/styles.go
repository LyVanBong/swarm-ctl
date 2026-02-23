package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	ColorPrimary  = lipgloss.Color("#7C3AED") // Purple
	ColorSuccess  = lipgloss.Color("#10B981") // Green
	ColorWarning  = lipgloss.Color("#F59E0B") // Amber
	ColorDanger   = lipgloss.Color("#EF4444") // Red
	ColorInfo     = lipgloss.Color("#3B82F6") // Blue
	ColorMuted    = lipgloss.Color("#6B7280") // Gray
	ColorHighlight = lipgloss.Color("#F3F4F6")

	// Base styles
	Bold   = lipgloss.NewStyle().Bold(true)
	Italic = lipgloss.NewStyle().Italic(true)
	Muted  = lipgloss.NewStyle().Foreground(ColorMuted)

	// Status styles
	Success = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	Danger  = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
	Info    = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)

	// Title / Header
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		PaddingTop(1).
		PaddingBottom(1)

	Subtitle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	// Banner box
	Banner = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 2).
		Bold(true).
		Foreground(ColorPrimary)

	// Section header
	SectionHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Background(ColorPrimary).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(1)

	// Table styles
	TableHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorMuted)

	TableRow = lipgloss.NewStyle().
		PaddingLeft(1)

	TableRowAlt = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(ColorMuted)

	// Status indicators
	StatusOK      = "✅"
	StatusFail    = "❌"
	StatusWarn    = "⚠️ "
	StatusRunning = "🔄"
	StatusPending = "⏳"

	// Box for info
	InfoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorInfo).
		Padding(0, 1)

	// Box for warning
	WarnBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(0, 1)

	// Box for error
	ErrorBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDanger).
		Padding(0, 1)

	// Box for success
	SuccessBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSuccess).
		Padding(0, 1)
)

// Render helpers
func RenderSuccess(msg string) string {
	return Success.Render("✅ " + msg)
}

func RenderError(msg string) string {
	return Danger.Render("❌ " + msg)
}

func RenderWarning(msg string) string {
	return Warning.Render("⚠️  " + msg)
}

func RenderInfo(msg string) string {
	return Info.Render("ℹ️  " + msg)
}

func RenderStep(step int, total int, msg string) string {
	counter := Muted.Render("[" + itoa(step) + "/" + itoa(total) + "]")
	return counter + " " + msg
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
