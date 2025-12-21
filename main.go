package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var engine = &Engine{}
var results = map[string]bool{}
var resultsMu = make(chan struct{}, 1)

func addSub(s string) {
	if s == "" {
		return
	}
	resultsMu <- struct{}{}
	results[s] = true
	<-resultsMu
}

func update(p *tea.Program, name string, count int, status string) {
	p.Send(Source{Name: name, Count: count, Status: status})
}

func printHelp() {
	fmt.Println("Usage: litocean -d domain.com | -l domains.txt")
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		return
	}

	ensureGo()
	ensureAllTools()

	d := flag.String("d", "", "domain")
	l := flag.String("l", "", "list")
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
			{"crt.sh", 0, "Waiting"},
			{"wayback", 0, "Waiting"},
			{"subfinder", 0, "Waiting"},
			{"amass", 0, "Waiting"},
			{"assetfinder", 0, "Waiting"},
			{"chaos", 0, "Waiting"},
			{"findomain", 0, "Waiting"},
		},
		StartTime: time.Now(),
	}

	p := tea.NewProgram(model)
	go p.Start()

	for _, d := range domains {
		go RunCrtSh(d, p)
		go RunWayback(d, p)
		go RunTool("subfinder", []string{"-d", d, "-all", "-silent"}, p)
		go RunTool("amass", []string{"enum", "-passive", "-d", d}, p)
		go RunTool("assetfinder", []string{"--subs-only", d}, p)
		go RunTool("chaos", []string{"-d", d, "-silent"}, p)
		go RunTool("findomain", []string{"-t", d, "-q"}, p)
	}

	select {}
}
