package main

import (
	"os"
	"os/exec"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	b, err := exec.Command("curl", "-H", `Authorization: token 3be14d5fad7e7889485c40987d5185c2d21a187b`, `-XGET`, `https://api.github.com/search/repositories?s=stars&o=desc&q=language%3Ago+stars%3A<5000&type=Repositories`).Output()
}
