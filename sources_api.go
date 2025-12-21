package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

/* ===================== HELPERS ===================== */

func updateSource(name string, count int, status string, p *tea.Program) {
	p.Send(Source{
		Name:   name,
		Count: count,
		Status: status,
	})
}

func addSub(sub string) {
	if strings.Contains(sub, "*") {
		sub = strings.ReplaceAll(sub, "*.", "")
	}
	resultsMu.Lock()
	results[sub] = true
	resultsMu.Unlock()
}

/* ===================== 1. ALIENVAULT ===================== */

func RunAlienVault(domain string, p *tea.Program) {
	updateSource("AlienVault", 0, "Running", p)

	url := fmt.Sprintf(
		"https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil {
		updateSource("AlienVault", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Passive []struct {
			Hostname string `json:"hostname"`
		} `json:"passive_dns"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	count := 0
	for _, h := range data.Passive {
		addSub(h.Hostname)
		count++
	}

	updateSource("AlienVault", count, "Done", p)
}

/* ===================== 2. CRT.SH ===================== */

func RunCrtSh(domain string, p *tea.Program) {
	updateSource("crt.sh", 0, "Running", p)

	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)
	resp, err := http.Get(url)
	if err != nil {
		updateSource("crt.sh", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	count := 0
	for _, entry := range data {
		names := strings.Split(entry["name_value"].(string), "\n")
		for _, n := range names {
			addSub(n)
			count++
		}
	}

	updateSource("crt.sh", count, "Done", p)
}

/* ===================== 3. CERTSPOTTER ===================== */

func RunCertSpotter(domain string, p *tea.Program) {
	updateSource("CertSpotter", 0, "Running", p)

	url := fmt.Sprintf(
		"https://api.certspotter.com/v1/issuances?domain=%s&include_subdomains=true&expand=dns_names",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil {
		updateSource("CertSpotter", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	var data []struct {
		DNS []string `json:"dns_names"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	count := 0
	for _, entry := range data {
		for _, s := range entry.DNS {
			addSub(s)
			count++
		}
	}

	updateSource("CertSpotter", count, "Done", p)
}

/* ===================== 4. HACKERTARGET ===================== */

func RunHackerTarget(domain string, p *tea.Program) {
	updateSource("HackerTarget", 0, "Running", p)

	url := fmt.Sprintf(
		"https://api.hackertarget.com/hostsearch/?q=%s",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil {
		updateSource("HackerTarget", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	count := 0
	for sc.Scan() {
		sub := strings.Split(sc.Text(), ",")[0]
		addSub(sub)
		count++
	}

	updateSource("HackerTarget", count, "Done", p)
}

/* ===================== 5. ANUBIS ===================== */

func RunAnubis(domain string, p *tea.Program) {
	updateSource("Anubis", 0, "Running", p)

	url := fmt.Sprintf(
		"https://jldc.me/anubis/subdomains/%s",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil {
		updateSource("Anubis", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	var subs []string
	json.NewDecoder(resp.Body).Decode(&subs)

	for _, s := range subs {
		addSub(s)
	}

	updateSource("Anubis", len(subs), "Done", p)
}

/* ===================== 6. URLSCAN ===================== */

func RunURLScan(domain string, p *tea.Program) {
	updateSource("URLScan", 0, "Running", p)

	url := fmt.Sprintf(
		"https://urlscan.io/api/v1/search/?q=domain:%s",
		domain,
	)

	resp, err := http.Get(url)
	if err != nil {
		updateSource("URLScan", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Results []struct {
			Page struct {
				Domain string `json:"domain"`
			} `json:"page"`
		} `json:"results"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	count := 0
	for _, r := range data.Results {
		addSub(r.Page.Domain)
		count++
	}

	updateSource("URLScan", count, "Done", p)
}

/* ===================== 7. VIRUSTOTAL (OPTIONAL KEY) ===================== */

func RunVirusTotal(domain string, p *tea.Program) {
	key := os.Getenv("VT_API_KEY")
	if key == "" {
		updateSource("VirusTotal", 0, "Skipped", p)
		return
	}

	updateSource("VirusTotal", 0, "Running", p)

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains?limit=1000", domain),
		nil,
	)
	req.Header.Set("x-apikey", key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		updateSource("VirusTotal", 0, "Error", p)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	for _, s := range data.Data {
		addSub(s.ID)
	}

	updateSource("VirusTotal", len(data.Data), "Done", p)
}
