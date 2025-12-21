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
)

var (
	results = make(map[string]bool)
	mu      sync.Mutex
)

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
		time.Sleep(2 * time.Millisecond)
	}
	fmt.Println("☠️🌊 LITOCEAN-GX 🌊☠️")
	fmt.Println("Ultimate Subdomain Enumeration Engine")
	fmt.Println("Developed by Biswajeet Ray")
	fmt.Println(strings.Repeat("=", 60))
}

func addResult(sub string) {
	mu.Lock()
	defer mu.Unlock()
	results[sub] = true
}

func count(label string, n int) {
	fmt.Printf("✅ %-15s → %d subdomains\n", label, n)
}

func fetchLines(url string, parser func(string) []string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	lines := parser(string(body))
	for _, s := range lines {
		addResult(s)
	}
	count(url, len(lines))
}

func alienvault(domain string) {
	url := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	items, ok := data["passive_dns"].([]interface{})
	if !ok {
		return
	}

	c := 0
	for _, v := range items {
		m := v.(map[string]interface{})
		host := m["hostname"].(string)
		addResult(host)
		c++
	}
	count("AlienVault", c)
}

func crtsh(domain string) {
	url := fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	c := 0
	for _, e := range data {
		names := strings.Split(e["name_value"].(string), "\n")
		for _, n := range names {
			n = strings.TrimPrefix(n, "*.")
			addResult(n)
			c++
		}
	}
	count("crt.sh", c)
}

func hackertarget(domain string) {
	url := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	c := 0
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ",")
		addResult(parts[0])
		c++
	}
	count("HackerTarget", c)
}

func runTool(name string, args ...string) {
	cmd := exec.Command(name, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	cmd.Start()

	sc := bufio.NewScanner(out)
	c := 0
	for sc.Scan() {
		addResult(sc.Text())
		c++
	}
	cmd.Wait()
	count(strings.ToUpper(name), c)
}

func install(tool string, cmd string) {
	if _, err := exec.LookPath(tool); err == nil {
		return
	}
	fmt.Println("⬇️ Installing", tool)
	exec.Command("bash", "-c", cmd).Run()
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "-d" {
		fmt.Println("Usage: LitOcean -d example.com")
		return
	}

	domain := os.Args[2]
	banner()

	fmt.Println("🔧 Checking & installing dependencies...")
	install("subfinder", "go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest")
	install("assetfinder", "go install github.com/tomnomnom/assetfinder@latest")
	install("amass", "sudo apt install -y amass")
	install("findomain", "curl -LO https://github.com/findomain/findomain/releases/latest/download/findomain-linux && chmod +x findomain-linux && sudo mv findomain-linux /usr/bin/findomain")

	fmt.Println("🚀 Starting enumeration...\n")

	var wg sync.WaitGroup

	wg.Add(3)
	go func() { defer wg.Done(); alienvault(domain) }()
	go func() { defer wg.Done(); crtsh(domain) }()
	go func() { defer wg.Done(); hackertarget(domain) }()
	wg.Wait()

	fmt.Println("\n⚙️ Running local tools...\n")

	runTool("subfinder", "-d", domain, "-silent")
	runTool("assetfinder", "--subs-only", domain)
	runTool("amass", "enum", "-passive", "-d", domain)
	runTool("findomain", "-t", domain, "-q")

	fmt.Println("\n📦 FINAL RESULTS")
	fmt.Println(strings.Repeat("-", 40))

	for s := range results {
		fmt.Println(s)
	}

	fmt.Printf("\n🔥 TOTAL UNIQUE SUBDOMAINS: %d\n", len(results))
}
