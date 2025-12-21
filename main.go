package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const author = "Biswajeet Ray"
const version = "1.0"

var domain string
var results = make(map[string][]string)
var mutex sync.Mutex

// ===================== BANNER & ANIMATION =====================

func banner() {
	fmt.Println(`
☠️🌊 LITOCEAN-GX 🌊☠️
Ultimate Subdomain Enumeration Engine
Developed by Biswajeet Ray
`)
}

func spinner(msg string, done chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r✔ %s completed\n", msg)
			return
		default:
			fmt.Printf("\r%s %s...", frames[i%len(frames)], msg)
			time.Sleep(120 * time.Millisecond)
			i++
		}
	}
}

// ===================== UTILITIES =====================

func unique(input []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range input {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func save(filename string, data []string) {
	f, _ := os.Create(filename)
	defer f.Close()
	for _, v := range data {
		fmt.Fprintln(f, v)
	}
}

// ===================== DEPENDENCY INSTALLER =====================

func installTool(name, pkg string) {
	if _, err := exec.LookPath(name); err == nil {
		return
	}
	fmt.Printf("[+] Installing %s\n", name)
	cmd := exec.Command("go", "install", pkg)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Run()
}

func installDependencies() {
	installTool("subfinder", "github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest")
	installTool("assetfinder", "github.com/tomnomnom/assetfinder@latest")
	installTool("amass", "github.com/owasp-amass/amass/v4/...@master")
	installTool("findomain", "github.com/findomain/findomain@latest")
	installTool("httpx", "github.com/projectdiscovery/httpx/cmd/httpx@latest")
}

// ===================== API SOURCES =====================

func crtsh() {
	url := fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data []map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)

	var subs []string
	for _, entry := range data {
		if name, ok := entry["name_value"].(string); ok {
			for _, s := range strings.Split(name, "\n") {
				s = strings.TrimPrefix(s, "*.")
				if strings.HasSuffix(s, domain) {
					subs = append(subs, s)
				}
			}
		}
	}

	mutex.Lock()
	results["crt.sh"] = unique(subs)
	mutex.Unlock()
}

func alienvault() {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	type Resp struct {
		Passive []struct {
			Hostname string `json:"hostname"`
		} `json:"passive_dns"`
	}

	var r Resp
	json.NewDecoder(resp.Body).Decode(&r)

	var subs []string
	for _, h := range r.Passive {
		if strings.HasSuffix(h.Hostname, domain) {
			subs = append(subs, h.Hostname)
		}
	}

	mutex.Lock()
	results["AlienVault"] = unique(subs)
	mutex.Unlock()
}

func anubis() {
	url := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var subs []string
	json.NewDecoder(resp.Body).Decode(&subs)

	mutex.Lock()
	results["Anubis"] = unique(subs)
	mutex.Unlock()
}

// ===================== TOOL SOURCES =====================

func runTool(name string, args []string) {
	cmd := exec.Command(name, args...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	var subs []string
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if strings.HasSuffix(s, domain) {
			subs = append(subs, s)
		}
	}
	cmd.Wait()

	mutex.Lock()
	results[name] = unique(subs)
	mutex.Unlock()
}

// ===================== MAIN =====================

func main() {
	flag.StringVar(&domain, "d", "", "Target domain")
	flag.Parse()

	if domain == "" {
		fmt.Println("Usage: ./litocean -d example.com")
		return
	}

	banner()
	fmt.Println("[*] Installing dependencies if missing...")
	installDependencies()

	var wg sync.WaitGroup

	// APIs
	apiTasks := []func(){crtsh, alienvault, anubis}
	for _, task := range apiTasks {
		wg.Add(1)
		go func(t func()) {
			done := make(chan bool)
			go spinner("API Source", done)
			t()
			done <- true
			wg.Done()
		}(task)
	}

	// Tools
	wg.Add(1)
	go func() {
		runTool("subfinder", []string{"-d", domain, "-silent"})
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runTool("assetfinder", []string{"--subs-only", domain})
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runTool("findomain", []string{"-t", domain, "-q"})
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runTool("amass", []string{"enum", "-passive", "-d", domain})
		wg.Done()
	}()

	wg.Wait()

	// Merge
	var all []string
	fmt.Println("\n📊 SOURCE STATISTICS")
	for src, subs := range results {
		fmt.Printf("✔ %-12s : %d\n", src, len(subs))
		all = append(all, subs...)
	}

	all = unique(all)
	save("subs.txt", all)

	fmt.Printf("\n☠️ TOTAL UNIQUE SUBDOMAINS: %d\n", len(all))

	// Alive check
	fmt.Println("\n🌐 Probing live hosts...")
	cmd := exec.Command("httpx", "-l", "subs.txt", "-silent")
	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")
	save("alive.txt", lines)

	fmt.Printf("✔ ALIVE HOSTS: %d\n", len(lines))
	fmt.Println("\n🔥 Recon Completed Successfully")
}
