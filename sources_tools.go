package main

import (
	"bufio"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func RunTool(name string, args []string, p *tea.Program) {
	updateSource(name, 0, "Running", p)

	cmd := exec.Command(name, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		updateSource(name, 0, "Error", p)
		return
	}
	cmd.Start()

	sc := bufio.NewScanner(out)
	count := 0
	for sc.Scan() {
		addSub(sc.Text())
		count++
	}

	cmd.Wait()
	updateSource(name, count, "Done", p)
}
