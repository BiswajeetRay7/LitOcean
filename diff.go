package main

import (
	"bufio"
	"os"
	"sort"
)

func loadPrevious(file string) map[string]bool {
	prev := map[string]bool{}
	f, err := os.Open(file)
	if err != nil {
		return prev
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		prev[sc.Text()] = true
	}
	return prev
}

func saveDiff(prev, current map[string]bool) {
	var diff []string
	for s := range current {
		if !prev[s] {
			diff = append(diff, s)
		}
	}
	if len(diff) == 0 {
		return
	}
	sort.Strings(diff)
	f, _ := os.Create("diff.txt")
	defer f.Close()
	for _, s := range diff {
		f.WriteString(s + "\n")
	}
}
