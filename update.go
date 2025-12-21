package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

func selfUpdate() {
	fmt.Println("🌊 Checking for updates...")
	api := "https://api.github.com/repos/BiswajeetRay7/LitOcean/releases/latest"

	resp, err := http.Get(api)
	if err != nil {
		fmt.Println("❌ Update check failed")
		return
	}
	defer resp.Body.Close()

	var data struct {
		Tag string `json:"tag_name"`
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	if data.Tag == Version {
		fmt.Println("✅ Already up to date")
		return
	}

	target := fmt.Sprintf("litocean-%s-%s", runtime.GOOS, runtime.GOARCH)
	var url string
	for _, a := range data.Assets {
		if a.Name == target {
			url = a.URL
		}
	}
	if url == "" {
		fmt.Println("❌ No compatible binary")
		return
	}

	tmp := "/tmp/litocean_update"
	out, _ := os.Create(tmp)
	resp, _ = http.Get(url)
	io.Copy(out, resp.Body)
	out.Close()
	os.Chmod(tmp, 0755)

	path, _ := exec.LookPath("litocean")
	os.Rename(tmp, path)
	fmt.Println("✅ Update complete")
}
