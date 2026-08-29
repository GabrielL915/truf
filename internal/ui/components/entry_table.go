package components

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

type EntryTable struct {
	Title      string
	Kind       ledger.Kind
	Entries    []ledger.Entry
	Categories []string
	Cursor     int
	Focused    bool
	Width      int
	Height     int

	Editing       bool
	EditingColumn int
	EditBuffer    string
	originalEntry *ledger.Entry

	colWidths []int
}

func NewEntryTable(title string, kind ledger.Kind) *EntryTable {
	return &EntryTable{
		Title:     title,
		Kind:      kind,
		Entries:   make([]ledger.Entry, 0),
		colWidths: []int{12, 30, 15, 12},
	}
}

func (t *EntryTable) isIncome() bool {
	return t.Kind == ledger.Income
}

func (t *EntryTable) SetEntries(entries []ledger.Entry) {
	t.Entries = entries
	if t.Cursor >= len(entries) {
		t.Cursor = len(entries) - 1
	}
	if t.Cursor < 0 {
		t.Cursor = 0
	}
}

func (t *EntryTable) SetCategories(categories []string) {
	t.Categories = categories
}

func (t *EntryTable) SelectByID(id string) {
	for i, e := range t.Entries {
		if e.ID == id {
			t.Cursor = i
			return
		}
	}
}

func (t *EntryTable) Current() (ledger.Entry, bool) {
	if t.Cursor < 0 || t.Cursor >= len(t.Entries) {
		return ledger.Entry{}, false
	}
	return t.Entries[t.Cursor], true
}

func (t *EntryTable) SetSize(width, height int) {
	t.Width = width
	t.Height = height

	available := width - 10
	t.colWidths = []int{12, available - 50, 18, 14}
	if t.colWidths[1] < 15 {
		t.colWidths[1] = 15
	}
}

func (t *EntryTable) Up() {
	if t.Cursor > 0 {
		t.Cursor--
	}
}

func (t *EntryTable) Down() {
	if t.Cursor < len(t.Entries)-1 {
		t.Cursor++
	}
}

func (t *EntryTable) StartEdit() {
	if len(t.Entries) == 0 {
		return
	}
	t.Editing = true
	t.EditingColumn = 0
	snapshot := t.Entries[t.Cursor]
	t.originalEntry = &snapshot
	t.EditBuffer = utils.FormatDate(snapshot.Date)
}

func (t *EntryTable) CancelEdit() {
	if t.originalEntry != nil && t.Cursor < len(t.Entries) {
		t.Entries[t.Cursor] = *t.originalEntry
	}
	t.originalEntry = nil
	t.Editing = false
	t.EditBuffer = ""
}

func (t *EntryTable) NextColumn() error {
	if !t.Editing {
		return nil
	}

	if err := t.saveCurrentColumn(); err != nil {
		return err
	}

	t.EditingColumn++
	if t.EditingColumn > 3 {
		t.Editing = false
		t.EditBuffer = ""
		t.originalEntry = nil
		return nil
	}

	t.loadCurrentColumn()
	return nil
}

func (t *EntryTable) saveCurrentColumn() error {
	if t.Cursor >= len(t.Entries) {
		return nil
	}
	entry := &t.Entries[t.Cursor]

	switch t.EditingColumn {
	case 0:
		date, err := utils.ParseDate(t.EditBuffer)
		if err != nil {
			return err
		}
		entry.Date = date
	case 1:
		entry.Description = t.EditBuffer
	case 2:
		entry.Category = t.EditBuffer
	case 3:
		amount, err := utils.ParseCurrency(t.EditBuffer)
		if err != nil {
			return err
		}
		entry.Amount = amount
	}
	return nil
}

func (t *EntryTable) loadCurrentColumn() {
	if t.Cursor >= len(t.Entries) {
		return
	}
	entry := t.Entries[t.Cursor]

	switch t.EditingColumn {
	case 0:
		t.EditBuffer = utils.FormatDate(entry.Date)
	case 1:
		t.EditBuffer = entry.Description
	case 2:
		t.EditBuffer = entry.Category
	case 3:
		t.EditBuffer = utils.FormatCurrency(entry.Amount)
	}
}

func (t *EntryTable) TypeChar(ch rune) {
	if !t.Editing {
		return
	}
	t.EditBuffer += string(ch)
}

func (t *EntryTable) Backspace() {
	if !t.Editing || len(t.EditBuffer) == 0 {
		return
	}
	_, size := utf8.DecodeLastRuneInString(t.EditBuffer)
	t.EditBuffer = t.EditBuffer[:len(t.EditBuffer)-size]
}

func (t *EntryTable) View() string {
	var sb strings.Builder

	titleStyle := styles.MenuTitleStyle
	if t.isIncome() {
		titleStyle = titleStyle.Foreground(styles.Success)
	} else {
		titleStyle = titleStyle.Foreground(styles.Danger)
	}
	sb.WriteString(titleStyle.Render(t.Title))
	sb.WriteString("\n\n")

	sb.WriteString(t.renderHeader())
	sb.WriteString("\n")

	totalWidth := 0
	for _, w := range t.colWidths {
		totalWidth += w + 1
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.Muted).Render(strings.Repeat("─", totalWidth)))
	sb.WriteString("\n")

	if len(t.Entries) == 0 {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true).
			Render("No entries. Press 'n' to add one."))
	} else {
		maxRows := t.Height - 10
		if maxRows < 5 {
			maxRows = 5
		}

		start := 0
		if t.Cursor >= maxRows {
			start = t.Cursor - maxRows + 1
		}

		end := start + maxRows
		if end > len(t.Entries) {
			end = len(t.Entries)
		}

		for i := start; i < end; i++ {
			sb.WriteString(t.renderRow(i))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	totalStyle := lipgloss.NewStyle().Bold(true)
	if t.isIncome() {
		totalStyle = totalStyle.Foreground(styles.Success)
	} else {
		totalStyle = totalStyle.Foreground(styles.Danger)
	}
	sb.WriteString(fmt.Sprintf("Total: %s", totalStyle.Render(utils.FormatCurrency(t.total()))))

	sb.WriteString("\n\n")
	sb.WriteString(styles.HelpStyle.Render("↑↓:navigate  Enter:edit  n:new  d:delete  Esc:back"))

	panelStyle := styles.PanelStyle
	if t.Focused {
		panelStyle = styles.FocusedPanelStyle
	}

	return panelStyle.
		Width(t.Width - 2).
		Height(t.Height - 2).
		Render(sb.String())
}

func (t *EntryTable) renderHeader() string {
	headers := []string{"Date", "Description", "Category", "Amount"}
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)

	var parts []string
	for i, h := range headers {
		parts = append(parts, headerStyle.Width(t.colWidths[i]).Render(h))
	}

	return strings.Join(parts, " ")
}

func (t *EntryTable) renderRow(idx int) string {
	entry := t.Entries[idx]
	isSelected := idx == t.Cursor

	cells := []string{
		utils.FormatDate(entry.Date),
		entry.Description,
		entry.Category,
		utils.FormatCurrency(entry.Amount),
	}

	var parts []string
	for col, value := range cells {
		editingCell := t.Editing && isSelected && t.EditingColumn == col
		if editingCell {
			value = t.EditBuffer + "▏"
		}

		style := t.getCellStyle(isSelected, editingCell)
		if col == 3 && !editingCell {
			if t.isIncome() {
				style = style.Foreground(styles.Success)
			} else {
				style = style.Foreground(styles.Danger)
			}
		}

		parts = append(parts, style.Width(t.colWidths[col]).Render(truncate(value, t.colWidths[col])))
	}

	return strings.Join(parts, " ")
}

func (t *EntryTable) getCellStyle(selected, editing bool) lipgloss.Style {
	if editing {
		return lipgloss.NewStyle().
			Background(styles.Primary).
			Foreground(styles.Background)
	}
	if selected {
		return lipgloss.NewStyle().
			Background(styles.Surface).
			Foreground(styles.Text)
	}
	return lipgloss.NewStyle().Foreground(styles.Text)
}

func (t *EntryTable) total() int64 {
	var total int64
	for _, e := range t.Entries {
		total += e.Amount
	}
	return total
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
