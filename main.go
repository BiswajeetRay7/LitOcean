package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

/* ===================== GLOBALS ===================== */

var (
	results = make(map[string]bool)
	mu      sync.Mutex
)

/* ===================== UTILS ===================== */

func add(sub string) {
	mu.Lock()
	results[sub] = true
	mu.Unlock()
}

/* ===================== TUI MODEL ===================== */

type source struct {
	Name     string
	Count    int
	Status   string
	Progress float64
}

type model struct {
	Sources []source
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	runStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

/* ===================== BANNER ===================== */

func banner() string {
	return `
██╗     ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗
██║     ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║     ██║   ██║   ██║   ██║██║     █████╗  ███████║██╔██╗ ██║
██║     ██║   ██║   ██║   ██║██║     ██╔══╝  ██╔══██║██║╚██╗██║
███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║
╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝

☠️🌊 LITOCEAN‑GX 🌊☠️
Developed by Biswajeet Ray
====================================================
`
}

/* ===================== INIT ===================== */

func initialModel() model {
	names := []string{
		"AlienVault", "crt.sh", "HackerTarget", "CertSpotter",
		"Anubis", "URLScan", "VirusTotal",
		"subfinder", "assetfinder", "amass", "findomain",
	}

	var srcs []source
	for _, n := range names {
		srcs = append(srcs, source{
			Name:     n,
			Status:   "Waiting",
			Progress: 0,
		})
	}
	return model{Sources: srcs}
}

/* ===================== TEA ===================== */

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s, ok := msg.(source); ok {
		for i := range m.Sources {
			if m.Sources[i].Name == s.Name {
				m.Sources[i] = s
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(banner())

	for _, s := range m.Sources {
		bar := progress.New(progress.WithDefaultGradient()).ViewAs(s.Progress)
		status := runStyle.Render(s.Status)
		if s.Status == "Done" {
			status = okStyle.Render("Done")
		}
		fmt.Fprintf(&b, "%-14s %s %4d  %s\n", s.Name, status, s.Count, bar)
	}

	fmt.Fprintf(&b, "\n🔥 TOTAL UNIQUE SUBDOMAINS: %d\n", len(results))
	b.WriteString("\nPress CTRL+C to exit\n")
	return b.String()
}

/* ===================== API FUNCTIONS ===================== */

func alienvault(domain string, ch chan source) {
	s := source{Name: "AlienVault", Status: "Running"}
	ch <- s

	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data map[string][]map[string]string
	json.NewDecoder(resp.Body).Decode(&data)

	for _, v := range data["passive_dns"] {
		add(v["hostname"])
		s.Count++
		s.Progress += 0.02
		ch <- s
	}
	s.Status = "Done"
	s.Progress = 1
	ch <- s
}

func crtsh(domain string, ch chan source) {
	s := source{Name: "crt.sh", Status: "Running"}
	ch <- s

	url := fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain)
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	for _, d := range data {
		names := strings.Split(d["name_value"].(string), "\n")
		for _, n := range names {
			add(strings.TrimPrefix(n, "*."))
			s.Count++
			s.Progress += 0.01
			ch <- s
		}
	}
	s.Status = "Done"
	s.Progress = 1
	ch <- s
}

/* ===================== LOCAL TOOLS ===================== */

func runTool(name string, args []string, ch chan source) {
	s := source{Name: name, Status: "Running"}
	ch <- s

	cmd := exec.Command(name, args...)
	out, _ := cmd.StdoutPipe()
	cmd.Start()

	sc := bufio.NewScanner(out)
	for sc.Scan() {
		add(sc.Text())
		s.Count++
		s.Progress += 0.02
		ch <- s
	}
	cmd.Wait()
	s.Status = "Done"
	s.Progress = 1
	ch <- s
}

/* ===================== MAIN ===================== */

func main() {
	if len(os.Args) < 3 || os.Args[1] != "-d" {
		fmt.Println("Usage: LitOcean -d example.com")
		return
	}

	domain := os.Args[2]
	ch := make(chan source)

	p := tea.NewProgram(initialModel())
	go p.Start()

	go alienvault(domain, ch)
	go crtsh(domain, ch)

	go runTool("subfinder", []string{"-d", domain, "-silent"}, ch)
	go runTool("assetfinder", []string{"--subs-only", domain}, ch)
	go runTool("amass", []string{"enum", "-passive", "-d", domain}, ch)
	go runTool("findomain", []string{"-t", domain, "-q"}, ch)

	for s := range ch {
		p.Send(s)
		time.Sleep(20 * time.Millisecond)
	}
}
