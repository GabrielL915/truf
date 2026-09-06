package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	frameWidth  = 100
	frameHeight = 30
)

func frame(m *Model) []string {
	return strings.Split(ansi.Strip(m.View()), "\n")
}

func frameText(m *Model) string {
	return ansi.Strip(m.View())
}

func sized(t *testing.T, m *Model) *Model {
	t.Helper()
	m.Update(tea.WindowSizeMsg{Width: frameWidth, Height: frameHeight})
	return m
}

func assertShape(t *testing.T, m *Model) []string {
	t.Helper()
	lines := frame(m)
	if len(lines) != frameHeight {
		t.Fatalf("frame has %d lines, want %d:\n%s", len(lines), frameHeight, strings.Join(lines, "\n"))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > frameWidth {
			t.Errorf("line %d is %d wide, want <= %d: %q", i, w, frameWidth, line)
		}
		if strings.ContainsAny(line, "╭╮╰╯┌┐└┘") {
			t.Errorf("line %d contains a box corner: %q", i, line)
		}
	}
	return lines
}

func TestShellFrameShapeOnEveryView(t *testing.T) {
	m := sized(t, monthModel(t))
	for _, open := range []func(){
		func() {},
		func() { key(m, "down"); key(m, "enter") },
		func() { key(m, "esc"); key(m, "down"); key(m, "down"); key(m, "enter") },
		func() { key(m, "esc"); key(m, "down"); key(m, "down"); key(m, "down"); key(m, "enter") },
		func() {
			key(m, "esc")
			key(m, "down")
			key(m, "down")
			key(m, "down")
			key(m, "down")
			key(m, "enter")
		},
	} {
		open()
		lines := assertShape(t, m)
		if !strings.Contains(lines[0], "TRUF") {
			t.Errorf("first line is not the tab bar: %q", lines[0])
		}
		if last := lines[len(lines)-1]; !strings.Contains(last, "esc") && !strings.Contains(last, "quit") {
			t.Errorf("last line is not the help bar: %q", last)
		}
	}
}

func TestShellHelpBarShowsStorageLabel(t *testing.T) {
	m := sized(t, monthModel(t))
	m.SetStorageLabel("~/.truf/truf.db")
	lines := frame(m)
	if last := lines[len(lines)-1]; !strings.HasSuffix(strings.TrimRight(last, " "), "~/.truf/truf.db") {
		t.Errorf("help bar lacks the storage label on the right: %q", last)
	}
}

func TestShellErrorReplacesHelpBarUntilNextKey(t *testing.T) {
	m, store := chaosModel(t)
	sized(t, m)
	store.SaveErr = errors.New("disk full")
	key(m, "down")
	key(m, "enter")
	key(m, "n")

	lines := frame(m)
	last := lines[len(lines)-1]
	if !strings.Contains(last, "Error: disk full") {
		t.Fatalf("help bar should show the error: %q", last)
	}
	if strings.Contains(last, "esc") {
		t.Errorf("help bar shows hints alongside the error: %q", last)
	}

	key(m, "esc")
	lines = frame(m)
	last = lines[len(lines)-1]
	if strings.Contains(last, "Error") {
		t.Errorf("error should clear on the next key: %q", last)
	}
	if !strings.Contains(last, "esc") {
		t.Errorf("hints should be back after the next key: %q", last)
	}
}

func TestShellTooSmallRendersGuardLine(t *testing.T) {
	m, _ := chaosModel(t)
	m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	out := frameText(m)
	if !strings.Contains(out, "80×24") {
		t.Fatalf("guard line missing:\n%s", out)
	}
	if strings.Contains(out, "Overview") {
		t.Errorf("too-small view still renders the shell:\n%s", out)
	}
	key(m, "down")
	key(m, "enter")
	if m.currentView != ViewOverview {
		t.Errorf("keys should be ignored while too small")
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if strings.Contains(frameText(m), "80×24") {
		t.Errorf("guard line still shown at 80×24")
	}
}
