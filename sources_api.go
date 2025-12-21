package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var results = map[string]bool{}
var resultsMu = &sync.Mutex{}

func addSub(s string) {
	s = strings.TrimPrefix(s, "*.")
	resultsMu.Lock()
	results[s] = true
	resultsMu.Unlock()
}

func update(p *tea.Program, name string, count int, status string) {
	p.Send(Source{Name: name, Count: count, Status: status})
}

func RunAlienVault(domain string, p *tea.Program) {
	update(p, "AlienVault", 0, "Running")

	url := fmt.Sprintf(
		"https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil {
		update(p, "AlienVault", 0, "Error")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Passive []struct {
			Hostname string `json:"hostname"`
		} `json:"passive_dns"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	for _, h := range data.Passive {
		addSub(h.Hostname)
	}

	update(p, "AlienVault", len(data.Passive), "Done")
}

func RunCrtSh(domain string, p *tea.Program) {
	update(p, "crt.sh", 0, "Running")

	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	count := 0
	for _, e := range data {
		for _, n := range strings.Split(e["name_value"].(string), "\n") {
			addSub(n)
			count++
		}
	}

	update(p, "crt.sh", count, "Done")
}

func RunHackerTarget(domain string, p *tea.Program) {
	update(p, "HackerTarget", 0, "Running")

	url := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	count := 0
	for sc.Scan() {
		addSub(strings.Split(sc.Text(), ",")[0])
		count++
	}

	update(p, "HackerTarget", count, "Done")
}
