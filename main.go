package main

import (
	"bufio"
	"flag"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var engine = &Engine{}

func main() {
	domain := flag.String("d", "", "Single domain")
	list := flag.String("l", "", "List of domains")
	flag.Parse()

	var domains []string
	if *domain != "" {
		domains = append(domains, *domain)
	}

	if *list != "" {
		f, _ := os.Open(*list)
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			domains = append(domains, sc.Text())
		}
	}

	model := Model{
		Sources: []Source{
			{"AlienVault", 0, "Waiting"},
			{"crt.sh", 0, "Waiting"},
			{"HackerTarget", 0, "Waiting"},
			{"subfinder", 0, "Waiting"},
			{"assetfinder", 0, "Waiting"},
			{"amass", 0, "Waiting"},
			{"findomain", 0, "Waiting"},
		},
	}

	p := tea.NewProgram(model)
	go p.Start()

	for _, d := range domains {
		go RunAlienVault(d, p)
		go RunCrtSh(d, p)
		go RunHackerTarget(d, p)

		go RunTool("subfinder", []string{"-d", d, "-silent"}, p)
		go RunTool("assetfinder", []string{"--subs-only", d}, p)
		go RunTool("amass", []string{"enum", "-passive", "-d", d}, p)
		go RunTool("findomain", []string{"-t", d, "-q"}, p)
	}

	select {}
}
