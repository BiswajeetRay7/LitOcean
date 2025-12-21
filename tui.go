package main

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

type Source struct {
	Name   string
	Count  int
	Status string
}

type Model struct {
	Sources []Source
	Cursor  int
}

func (m Model) Init() tea.Cmd { return nil }

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
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	out := `
☠️🌊 LITOCEAN-GX 🌊☠️
Subdomain Enumeration Engine
Developed by Biswajeet Ray
--------------------------------------------------
↑↓ scroll | s sort | p pause | r resume | q quit

`
	for i, s := range m.Sources {
		cursor := " "
		if i == m.Cursor {
			cursor = "➤"
		}
		out += fmt.Sprintf("%s %-15s %-8s %d\n",
			cursor, s.Name, s.Status, s.Count)
	}

	out += fmt.Sprintf("\n🔥 TOTAL UNIQUE SUBDOMAINS: %d\n", len(results))
	return out
}
