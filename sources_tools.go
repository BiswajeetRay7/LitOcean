package main

import (
	"bufio"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func RunTool(name string, args []string, p *tea.Program) {
	update(p, name, 0, "Running")

	cmd := exec.Command(name, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		update(p, name, 0, "Error")
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
	update(p, name, count, "Done")
}
