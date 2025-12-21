package main

import (
	"fmt"
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
	StartTime time.Time
	ShowHelp  bool
}

func stars(s string) string {
	if s == "Done" {
		return "⭐⭐⭐"
	}
	if s == "Running" {
		return "⭐⭐☆"
	}
	return "⭐☆☆"
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return t })
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
		case "h":
			m.ShowHelp = !m.ShowHelp
		case "e":
			exportResults()
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.ShowHelp {
		return `
🌊⭐ LITOCEAN HELP ⭐🌊

USAGE:
  litocean -d domain.com
  litocean -l domains.txt

KEYS:
  h  help
  e  export
  q  quit
`
	}

	title := lipgloss.NewStyle().Foreground(lipgloss.Color("#5FD7FF")).Bold(true)

	out := title.Render(`
██╗     ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗
██║     ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║     ██║   ██║   ██║   ██║██║     █████╗  ███████║██╔██╗ ██║
██║     ██║   ██║   ██║   ██║██║     ██╔══╝  ██╔══██║██║╚██╗██║
███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║
╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝

🌊⭐ LITOCEAN ⭐🌊
Subdomain Enumeration Framework
`)

	out += fmt.Sprintf("⏱ Elapsed: %s\n\n",
		time.Since(m.StartTime).Truncate(time.Second))

	for _, s := range m.Sources {
		out += fmt.Sprintf("%-12s %s %s %d\n",
			s.Name, stars(s.Status), s.Status, s.Count)
	}
	return out
}
