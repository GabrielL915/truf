package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func quits(m *Model, k string) bool {
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestTabBarListsViewsWithNumbersAndMonth(t *testing.T) {
	m := sized(t, monthModel(t))
	top := frame(m)[0]
	for _, want := range []string{"TRUF", "1 Overview", "2 Income", "3 Expenses", "4 Categories", "5 Settings", "‹ Mar 2026 ›"} {
		if !strings.Contains(top, want) {
			t.Errorf("tab bar lacks %q: %q", want, top)
		}
	}
}

func TestTabBarNumberKeysTabAndEscLandOnViews(t *testing.T) {
	m := sized(t, monthModel(t))
	for k, want := range map[string]ViewType{"1": ViewOverview, "2": ViewIncome, "3": ViewExpenses, "4": ViewCategories, "5": ViewSettings} {
		key(m, k)
		if m.currentView != want {
			t.Errorf("key %s landed on view %d, want %d", k, m.currentView, want)
		}
		if !strings.Contains(frameText(m), viewLabels[want]) {
			t.Errorf("frame after %s lacks label %q", k, viewLabels[want])
		}
	}

	key(m, "1")
	key(m, "tab")
	if m.currentView != ViewIncome {
		t.Errorf("tab from Overview = %d, want Income", m.currentView)
	}
	key(m, "shift+tab")
	key(m, "shift+tab")
	if m.currentView != ViewSettings {
		t.Errorf("shift+tab twice from Income = %d, want Settings (wraps)", m.currentView)
	}
	key(m, "tab")
	if m.currentView != ViewOverview {
		t.Errorf("tab from Settings = %d, want Overview (wraps)", m.currentView)
	}
	key(m, "4")
	key(m, "esc")
	if m.currentView != ViewOverview {
		t.Errorf("esc from Categories = %d, want Overview", m.currentView)
	}
}

func TestTabBarQuitOnlyOnOverview(t *testing.T) {
	m := sized(t, monthModel(t))
	key(m, "2")
	if quits(m, "q") {
		t.Error("q on Income should not quit")
	}
	key(m, "esc")
	if !quits(m, "q") {
		t.Error("q on Overview should quit")
	}
}

func TestTabBarMonthKeysWorkOnEveryView(t *testing.T) {
	m := sized(t, monthModel(t))
	for _, k := range []string{"1", "2", "3", "4", "5"} {
		key(m, k)
		key(m, "[")
		if top := frame(m)[0]; !strings.Contains(top, "Feb 2026") {
			t.Errorf("view %s: [ did not move the tab bar month to Feb 2026: %q", k, top)
		}
		key(m, "]")
		if top := frame(m)[0]; !strings.Contains(top, "Mar 2026") {
			t.Errorf("view %s: ] did not move the tab bar month back to Mar 2026: %q", k, top)
		}
	}
	key(m, "h")
	key(m, "h")
	key(m, "l")
	if top := frame(m)[0]; !strings.Contains(top, "Feb 2026") {
		t.Errorf("h h l should land on Feb 2026: %q", top)
	}
}

func TestTabBarBracketWhileEditingIsTyped(t *testing.T) {
	m := sized(t, monthModel(t))
	key(m, "2")
	key(m, "enter")
	key(m, "tab")
	typeText(m, "[")
	if top := frame(m)[0]; !strings.Contains(top, "Mar 2026") {
		t.Errorf("[ while editing changed the month: %q", top)
	}
	if !strings.Contains(m.incomeTable.EditBuffer, "[") {
		t.Errorf("[ while editing should be typed, buffer=%q", m.incomeTable.EditBuffer)
	}
	if last := frame(m)[frameHeight-1]; !strings.Contains(last, "cancel") {
		t.Errorf("help bar should show editing hints: %q", last)
	}
}

func TestTabBarFitsAtMinimumWidth(t *testing.T) {
	m, _ := chaosModel(t)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	top := frame(m)[0]
	if !strings.Contains(top, "5 Settings") || !strings.Contains(top, "Mar 26") {
		t.Errorf("tab bar at 80 columns should keep every tab and a short month: %q", top)
	}
}
