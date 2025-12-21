package main

import (
	"bufio"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func RunTool(name string, args []string, p *tea.Program) {
	update(p, name, 0, "Running")

	cmd := exec.Command(name, args...)
	out, _ := cmd.StdoutPipe()
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
