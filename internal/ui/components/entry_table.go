package components

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/gabriel-luiz/truf/internal/ledger"
	"github.com/gabriel-luiz/truf/internal/ui/styles"
	"github.com/gabriel-luiz/truf/pkg/utils"
)

const (
	tableChromeLines = 5
	dateWidth        = 5
	categoryWidth    = 18
	amountWidth      = 14
	cellGap          = 2
	distributionTop  = 5
)

type EntryTable struct {
	Title      string
	Kind       ledger.Kind
	Month      time.Time
	Entries    []ledger.Entry
	Categories []string
	Cursor     int
	Width      int
	Height     int

	Editing       bool
	EditingColumn int
	EditBuffer    string
	originalEntry *ledger.Entry
}

func NewEntryTable(title string, kind ledger.Kind) *EntryTable {
	return &EntryTable{
		Title:   title,
		Kind:    kind,
		Entries: make([]ledger.Entry, 0),
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

func (t *EntryTable) SetMonth(month time.Time) {
	t.Month = month
}

func (t *EntryTable) ResetCursor() {
	t.Cursor = 0
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
	lines := make([]string, 0, t.Height)
	lines = append(lines, t.renderTitleLine(), t.renderColumnHeaders(), t.rule())

	rows := max(t.Height-tableChromeLines, 1)
	if len(t.Entries) == 0 {
		lines = append(lines, " "+styles.MutedStyle.Render(
			fmt.Sprintf("No entries in %s. Press n to add one.", utils.FormatMonthYearFull(t.Month))))
	} else {
		start, end := t.window(rows)
		for i := start; i < end; i++ {
			lines = append(lines, t.renderRow(i))
		}
		for len(lines) < t.Height-2 {
			lines = append(lines, "")
		}
		lines = append(lines, t.rule(), t.renderFooter(start, end))
	}

	return lipgloss.NewStyle().
		Width(t.Width).
		Height(t.Height).
		MaxHeight(t.Height).
		Render(strings.Join(lines, "\n"))
}

func (t *EntryTable) window(rows int) (start, end int) {
	if t.Cursor >= rows {
		start = t.Cursor - rows + 1
	}
	end = min(start+rows, len(t.Entries))
	return start, end
}

func (t *EntryTable) columnWidths() (date, description, category, amount int) {
	fixed := 2 + dateWidth + cellGap + cellGap + categoryWidth + cellGap + amountWidth + 1
	return dateWidth, max(t.Width-fixed, 10), categoryWidth, amountWidth
}

func (t *EntryTable) renderTitleLine() string {
	left := lipgloss.NewStyle().Bold(true).Foreground(styles.KindColor(t.isIncome())).Render(t.Title) +
		styles.MutedStyle.Render(fmt.Sprintf(" · %s · %s", entries(len(t.Entries)), utils.FormatMonthYearFull(t.Month)))
	right := styles.MutedStyle.Render("Total ") +
		lipgloss.NewStyle().Bold(true).Foreground(styles.KindColor(t.isIncome())).Render(utils.FormatCurrency(t.total()))
	gap := max(t.Width-2-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return " " + left + strings.Repeat(" ", gap) + right
}

func (t *EntryTable) renderColumnHeaders() string {
	date, description, category, amount := t.columnWidths()
	gap := strings.Repeat(" ", cellGap)
	h := styles.MutedStyle
	return "  " +
		h.Width(date).Render("DATE") + gap +
		h.Width(description).Render("DESCRIPTION") + gap +
		h.Width(category).Render("CATEGORY") + gap +
		h.Width(amount).Align(lipgloss.Right).Render("AMOUNT")
}

func (t *EntryTable) rule() string {
	return " " + styles.ChromeStyle.Render(strings.Repeat("─", max(t.Width-2, 1)))
}

func (t *EntryTable) renderRow(idx int) string {
	entry := t.Entries[idx]
	selected := idx == t.Cursor
	editing := t.Editing && selected
	date, description, category, amount := t.columnWidths()
	widths := []int{date, description, category, amount}

	cells := []string{
		entry.Date.Format("02/01"),
		entry.Description,
		entry.Category,
		utils.FormatCurrency(entry.Amount),
	}

	base := styles.TextStyle
	marker := "  "
	switch {
	case editing:
		base = styles.MutedStyle
		marker = styles.AccentStyle.Render("▌") + " "
	case selected:
		base = lipgloss.NewStyle().Background(styles.Surface).Foreground(styles.Text)
		marker = styles.AccentStyle.Background(styles.Surface).Render("▌") + base.Render(" ")
	}

	parts := make([]string, 0, len(cells))
	for col, value := range cells {
		style := base
		if editing && t.EditingColumn == col {
			value = t.EditBuffer + "▏"
			style = lipgloss.NewStyle().Background(styles.Accent).Foreground(styles.OnAccent)
		} else if col == 3 {
			style = style.Foreground(styles.KindColor(t.isIncome()))
		}
		if col == 3 {
			style = style.Align(lipgloss.Right)
		}
		parts = append(parts, style.Width(widths[col]).Render(ansi.Truncate(value, widths[col], "…")))
	}

	row := marker + strings.Join(parts, base.Render(strings.Repeat(" ", cellGap)))
	if pad := t.Width - lipgloss.Width(row); pad > 0 {
		row += base.Render(strings.Repeat(" ", pad))
	}
	return row
}

func (t *EntryTable) renderFooter(start, end int) string {
	right := ""
	if len(t.Entries) > end-start {
		right = styles.MutedStyle.Render(fmt.Sprintf("%d–%d of %d", start+1, end, len(t.Entries)))
	}
	room := t.Width - 2 - lipgloss.Width(right)
	if right != "" {
		room -= cellGap
	}
	left := styles.MutedStyle.Render(ansi.Truncate(t.distribution(), max(room, 0), "…"))
	gap := max(t.Width-2-lipgloss.Width(left)-lipgloss.Width(right), 0)
	return " " + left + strings.Repeat(" ", gap) + right
}

type share struct {
	name   string
	amount int64
}

func (t *EntryTable) distribution() string {
	total := t.total()
	if total == 0 {
		return ""
	}
	byCategory := map[string]int64{}
	for _, e := range t.Entries {
		name := e.Category
		if name == "" {
			name = "Uncategorised"
		}
		byCategory[name] += e.Amount
	}
	shares := make([]share, 0, len(byCategory))
	for name, amount := range byCategory {
		shares = append(shares, share{name, amount})
	}
	sort.Slice(shares, func(i, j int) bool {
		if shares[i].amount != shares[j].amount {
			return shares[i].amount > shares[j].amount
		}
		return shares[i].name < shares[j].name
	})
	if len(shares) > distributionTop+1 {
		var other int64
		for _, s := range shares[distributionTop:] {
			other += s.amount
		}
		shares = append(shares[:distributionTop], share{"Other", other})
	}
	parts := make([]string, 0, len(shares))
	for _, s := range shares {
		parts = append(parts, fmt.Sprintf("%s %d%%", s.name, (s.amount*100+total/2)/total))
	}
	return strings.Join(parts, " · ")
}

func (t *EntryTable) total() int64 {
	var total int64
	for _, e := range t.Entries {
		total += e.Amount
	}
	return total
}
