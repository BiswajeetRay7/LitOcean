package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
)

type source struct {
	name   string
	count  int
	status string
}

type model struct {
	sources []source
	total   int
	progressBars map[string]progress.Model
}

func initialModel() model {
	srcs := []source{
		{"AlienVault", 0, "Pending"},
		{"crt.sh", 0, "Pending"},
		{"HackerTarget", 0, "Pending"},
		{"Subfinder", 0, "Pending"},
		{"Assetfinder", 0, "Pending"},
		{"Amass", 0, "Pending"},
		{"Findomain", 0, "Pending"},
	}
	pBars := make(map[string]progress.Model)
	for _, s := range srcs {
		pBars[s.name] = progress.New(progress.WithDefaultGradient())
	}
	return model{
		sources:      srcs,
		progressBars: pBars,
		total:        0,
	}
}

func banner() {
	ascii := `
██╗     ██╗████████╗ ██████╗  ██████╗███████╗ █████╗ ███╗   ██╗
██║     ██║╚══██╔══╝██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
██║     ██║   ██║   ██║   ██║██║     █████╗  ███████║██╔██╗ ██║
██║     ██║   ██║   ██║   ██║██║     ██╔══╝  ██╔══██║██║╚██╗██║
███████╗██║   ██║   ╚██████╔╝╚██████╗███████╗██║  ██║██║ ╚████║
╚══════╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝
`
	for _, c := range ascii {
		fmt.Print(string(c))
		time.Sleep(1 * time.Millisecond)
	}
	fmt.Println("\n☠️🌊 LITOCEAN-GX 🌊☠️")
	fmt.Println("Developed by Biswajeet Ray")
	fmt.Println(strings.Repeat("=", 60))
}

// helper to add unique subdomain
var mu sync.Mutex
var results = make(map[string]bool)

func addResult(sub string) {
	mu.Lock()
	defer mu.Unlock()
	results[sub] = true
}

// fetch API and update counts live
func fetchAlienvault(domain string, ch chan source) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)
	resp, err := http.Get(url)
	if err != nil {
		ch <- source{"AlienVault", 0, "Error"}
		return
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	items, ok := data["passive_dns"].([]interface{})
	c := 0
	if ok {
		for _, v := range items {
			host := v.(map[string]interface{})["hostname"].(string)
			addResult(host)
			c++
		}
	}
	ch <- source{"AlienVault", c, "Done"}
}

// helper to run local CLI tool
func runTool(name string, args ...string) source {
	cmd := exec.Command(name, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return source{name, 0, "Error"}
	}
	cmd.Start()
	sc := bufio.NewScanner(out)
	c := 0
	for sc.Scan() {
		addResult(sc.Text())
		c++
	}
	cmd.Wait()
	if err != nil {
		return source{name, c, "Error"}
	}
	return source{name, c, "Done"}
}

// BubbleTea update loop
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) View() string {
	s := bannerText()
	for _, src := range m.sources {
		bar := m.progressBars[src.name].View()
		s += fmt.Sprintf("%-15s [%s] → %d subdomains\n", src.name, src.status, src.count)
		s += bar + "\n"
	}
	s += fmt.Sprintf("\n🔥 TOTAL UNIQUE SUBDOMAINS: %d\n", len(results))
	return s
}

func bannerText() string {
	return `
☠️🌊 LITOCEAN-GX 🌊☠️
Developed by Biswajeet Ray
============================================================
`
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "-d" {
		fmt.Println("Usage: LitOcean -d example.com")
		return
	}
	domain := os.Args[2]

	// initial banner
	banner()

	// create BubbleTea program
	p := tea.NewProgram(initialModel())
	ch := make(chan source)

	// start fetching APIs
	go fetchAlienvault(domain, ch)
	// similarly add crt.sh, HackerTarget, etc in parallel...

	// run local tools in parallel
	var wg sync.WaitGroup
	localTools := []struct {
		name string
		args []string
	}{
		{"subfinder", []string{"-d", domain, "-silent"}},
		{"assetfinder", []string{"--subs-only", domain}},
		{"amass", []string{"enum", "-passive", "-d", domain}},
		{"findomain", []string{"-t", domain, "-q"}},
	}

	for _, t := range localTools {
		wg.Add(1)
		go func(tool string, args []string) {
			defer wg.Done()
			res := runTool(tool, args...)
			ch <- res
		}(t.name, t.args)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	// update BubbleTea model from channel
	go func() {
		m := initialModel()
		for s := range ch {
			for i := range m.sources {
				if m.sources[i].name == s.name {
					m.sources[i] = s
				}
			}
			m.total = len(results)
			p.Send(m)
		}
	}()

	p.Start()
}
