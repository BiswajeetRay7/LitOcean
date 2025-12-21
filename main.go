package main

import (
	"bufio"
	"flag"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var engine = &Engine{}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "update" {
		selfUpdate()
		return
	}

	d := flag.String("d", "", "Domain")
	l := flag.String("l", "", "List")
	flag.Parse()

	var domains []string
	if *d != "" {
		domains = append(domains, *d)
	}
	if *l != "" {
		f, _ := os.Open(*l)
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
		StartTime: time.Now(),
		Dark:      true,
	}

	p := tea.NewProgram(model)
	go p.Start()

	for _, dom := range domains {
		go RunAlienVault(dom, p)
		go RunCrtSh(dom, p)
		go RunHackerTarget(dom, p)

		go RunTool("subfinder", []string{"-d", dom, "-silent"}, p)
		go RunTool("assetfinder", []string{"--subs-only", dom}, p)
		go RunTool("amass", []string{"enum", "-passive", "-d", dom}, p)
		go RunTool("findomain", []string{"-t", dom, "-q"}, p)
	}

	select {}
}
