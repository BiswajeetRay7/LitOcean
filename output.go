package main

import (
	"os"
	"os/exec"
	"sort"
)

func exportResults() {
	resultsMu.Lock()
	defer resultsMu.Unlock()

	var subs []string
	for s := range results {
		subs = append(subs, s)
	}
	sort.Strings(subs)

	f, _ := os.Create("subs.txt")
	for _, s := range subs {
		f.WriteString(s + "\n")
	}
	f.Close()

	exec.Command("httpx", "-l", "subs.txt", "-silent",
		"-threads", "200", "-o", "alive.txt").Run()
}
