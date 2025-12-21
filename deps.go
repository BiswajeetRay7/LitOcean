package main

import "os/exec"

func ensureHTTPX() {
	if _, err := exec.LookPath("httpx"); err == nil {
		return
	}
	exec.Command(
		"go", "install",
		"github.com/projectdiscovery/httpx/cmd/httpx@latest",
	).Run()
}
