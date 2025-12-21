package main

import (
	"os"
	"os/exec"
	"sort"
)

func saveSubs(file string) {
	resultsMu.Lock()
	defer resultsMu.Unlock()

	var subs []string
	for s := range results {
		subs = append(subs, s)
	}
	sort.Strings(subs)

	f, _ := os.Create(file)
	defer f.Close()

	for _, s := range subs {
		f.WriteString(s + "\n")
	}
}

func exportAll() {
	prev := loadPrevious("subs.txt")
	saveSubs("subs.txt")
	saveDiff(prev, results)
	ensureHTTPX()

	exec.Command(
		"httpx",
		"-l", "subs.txt",
		"-silent",
		"-sc",
		"-title",
		"-o", "alive.txt",
	).Run()
}
