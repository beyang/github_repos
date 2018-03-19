package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ReposResult struct {
	Items []Repo `json:"items,omitempty"`
}

type Repo struct {
	FullName string `json:"full_name,omitempty"`
}

func run() {
	langs := []string{"go", "java", "python"}
	langChans := map[string]chan struct{}{
		"go":     make(chan struct{}, 10),
		"java":   make(chan struct{}, 10),
		"python": make(chan struct{}, 10),
	}
	for _, lang := range langs {
		go func(lang string) {
			for i := 0; i < 100; i++ {
				go func(i int) {
					langChans[lang] <- struct{}{}
					defer func() { <-langChans[lang] }()

					start, end := i*100+1, (i+1)*100
					outfile := fmt.Sprintf("%s_%d-%d.txt", lang, start, end)
					if _, err := os.Stat(outfile); err == nil { // Skip files that already exist
						return
					}

					log.Printf("Querying %s, %d - %d", lang, start, end)
					args := []string{"-H", `Authorization: token 3be14d5fad7e7889485c40987d5185c2d21a187b`, `-XGET`,
						`https://api.github.com/search/repositories?s=stars&o=desc&q=language%3A` + lang + `+stars%3A<5000&type=Repositories&per_page=100&page=` + strconv.Itoa(i+1)}
					b, err := exec.Command("curl", args...).Output()
					if err != nil {
						log.Fatal(err)
					}

					var r ReposResult
					if err := json.Unmarshal(b, &r); err != nil {
						log.Fatal(err)
					}

					repoNames := make([]string, 0, 100)
					for _, item := range r.Items {
						repoNames = append(repoNames, item.FullName)
					}
					if err := ioutil.WriteFile(outfile, []byte(strings.Join(repoNames, "\n")), 0644); err != nil {
						log.Fatal(err)
					}
					log.Printf("...Queried %s, %d - %d", lang, start, end)
				}(i)
			}
		}(lang)
	}
	select {}
}

func main() {
	run()
}
