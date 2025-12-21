package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func RunCrtSh(domain string, p *tea.Program) {
	update(p, "crt.sh", 0, "Running")

	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
	resp, err := http.Get(url)
	if err != nil || resp == nil {
		update(p, "crt.sh", 0, "Error")
		return
	}
	defer resp.Body.Close()

	var data []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		update(p, "crt.sh", 0, "Error")
		return
	}

	count := 0
	for _, e := range data {
		if v, ok := e["name_value"].(string); ok {
			for _, s := range strings.Split(v, "\n") {
				addSub(s)
				count++
			}
		}
	}
	update(p, "crt.sh", count, "Done")
}

func RunWayback(domain string, p *tea.Program) {
	update(p, "wayback", 0, "Running")

	url := fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&fl=original&collapse=urlkey",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil || resp == nil {
		update(p, "wayback", 0, "Error")
		return
	}
	defer resp.Body.Close()

	var rows [][]string
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		update(p, "wayback", 0, "Error")
		return
	}

	count := 0
	for _, r := range rows {
		if len(r) > 0 {
			host := strings.Split(strings.TrimPrefix(r[0], "http"), "/")[0]
			host = strings.TrimPrefix(host, "s://")
			host = strings.TrimPrefix(host, "://")
			addSub(host)
			count++
		}
	}
	update(p, "wayback", count, "Done")
}
