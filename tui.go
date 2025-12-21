package main

import (
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Source struct {
	Name   string
	Count  int
	Status string
}

type Model struct {
	Sources   []Source
	Cursor    int
	StartTime time.Time
}

func stars(status string) string {
	switch status {
	case "Done":
		return "⭐⭐⭐"
	case "Running":
		return "⭐⭐☆"
	default:
		return "⭐☆☆"
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case Source:
		for i := range m.Sources {
			if m.Sources[i].Name == msg.Name {
				m.Sources[i] = msg
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down":
			if m.Cursor < len(m.Sources)-1 {
				m.Cursor++
			}
		case "s":
			sort.Slice(m.Sources, func(i, j int) bool {
				return m.Sources[i].Count > m.Sources[j].Count
			})
		case "p":
			engine.Pause()
		case "r":
			engine.Resume()
		case "e":
			exportAll()
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FD7FF")).
		Bold(true)

	out := title.Render("🌊⭐ LITOCEAN ⭐🌊\nAdvanced Subdomain Enumeration Framework\n")
	out += fmt.Sprintf("⏱ Elapsed: %s\n\n", time.Since(m.StartTime).Truncate(time.Second))

	for i, s := range m.Sources {
		cursor := " "
		if i == m.Cursor {
			cursor = "➤"
		}
		out += fmt.Sprintf(
			"%s %-14s %s %-8s %d\n",
			cursor,
			s.Name,
			stars(s.Status),
			s.Status,
			s.Count,
		)
	}

	out += fmt.Sprintf("\n🔥 TOTAL UNIQUE SUBDOMAINS: %d\n", len(results))
	out += "\nKeys: ↑↓ scroll | s sort | e export | p pause | r resume | q quit\n"
	return out
}
