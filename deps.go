package main

import (
	"fmt"
	"os"
	"os/exec"
)

func ensureGo() {
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("❌ Go is required but not installed")
		os.Exit(1)
	}
}

func ensureTool(name string, install func()) {
	if _, err := exec.LookPath(name); err == nil {
		return
	}
	fmt.Printf("⬇️ Installing %s...\n", name)
	install()
}

func ensureAllTools() {
	ensureTool("subfinder", func() {
		exec.Command("go", "install",
			"github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest").Run()
	})
	ensureTool("amass", func() {
		exec.Command("go", "install",
			"github.com/owasp-amass/amass/v4/cmd/amass@latest").Run()
	})
	ensureTool("assetfinder", func() {
		exec.Command("go", "install",
			"github.com/tomnomnom/assetfinder@latest").Run()
	})
	ensureTool("chaos", func() {
		exec.Command("go", "install",
			"github.com/projectdiscovery/chaos-client/cmd/chaos@latest").Run()
	})
	ensureTool("findomain", func() {
		exec.Command("bash", "-c",
			"curl -sL https://github.com/findomain/findomain/releases/latest/download/findomain-linux -o /tmp/findomain").Run()
		exec.Command("chmod", "+x", "/tmp/findomain").Run()
		exec.Command("sudo", "mv", "/tmp/findomain", "/usr/local/bin/findomain").Run()
	})
	ensureTool("httpx", func() {
		exec.Command("go", "install",
			"github.com/projectdiscovery/httpx/cmd/httpx@latest").Run()
	})
	ensureTool("anew", func() {
		exec.Command("go", "install",
			"github.com/tomnomnom/anew@latest").Run()
	})
}
