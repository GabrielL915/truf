package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
)

type Hint struct {
	Key  string
	Verb string
}

type HelpBar struct {
	Hints        []Hint
	StorageLabel string
	Err          error
	Width        int
}

func NewHelpBar() *HelpBar {
	return &HelpBar{}
}

func (h *HelpBar) SetHints(hints []Hint) {
	h.Hints = hints
}

func (h *HelpBar) SetError(err error) {
	h.Err = err
}

func (h *HelpBar) SetWidth(width int) {
	h.Width = width
}

func (h *HelpBar) View() string {
	if h.Err != nil {
		return styles.BarStyle.Width(h.Width).Render(
			styles.DangerStyle.Background(styles.Surface).Render(" Error: " + h.Err.Error()))
	}

	parts := make([]string, 0, len(h.Hints))
	for _, hint := range h.Hints {
		parts = append(parts, styles.KeyStyle.Render(hint.Key)+styles.BarStyle.Render(" "+hint.Verb))
	}
	left := " " + strings.Join(parts, styles.BarStyle.Render("  "))

	right := ""
	if h.StorageLabel != "" {
		right = styles.BarStyle.Foreground(styles.Muted).Render(h.StorageLabel) + " "
	}

	gap := h.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		right = ""
		gap = h.Width - lipgloss.Width(left)
	}
	if gap < 0 {
		left = ansi.Truncate(left, h.Width, "")
		gap = 0
	}
	return styles.BarStyle.Width(h.Width).Render(left + styles.BarStyle.Render(strings.Repeat(" ", gap)) + right)
}
