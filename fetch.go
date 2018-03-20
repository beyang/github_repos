package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ReposResult struct {
	Items []Repo `json:"items,omitempty"`
}

type Repo struct {
	Stars    int    `json:"stargazers_count"`
	FullName string `json:"full_name,omitempty"`
}

var limiter = make(chan struct{}, 10)

func init() {
	go func() {
		for {
			limiter <- struct{}{}
			time.Sleep(time.Second)
		}
	}()
}

func run() {
	var wg sync.WaitGroup
	langs := []string{"go", "java", "python"}
	wg.Add(len(langs))
	for _, lang := range langs {
		go func(lang string) {
			defer wg.Done()

			maxStars := 5000
			for i := 0; i < 10; i++ {
				maxStars = traunch(lang, maxStars, i)
				if maxStars == 0 {
					break
				}
			}
		}(lang)
	}
	wg.Wait()
}

func traunch(lang string, maxStars int, iter int) (nextMaxStars int) {
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ { // get 1000
		go func(i int) {
			defer wg.Done()
			<-limiter

			start, end := (iter*1000)+i*100+1, (iter*1000)+(i+1)*100
			outfile := fmt.Sprintf("data/%s_%d-%d.txt", lang, start, end)

			log.Printf("Querying %s, %d - %d", lang, start, end)
			args := []string{"-H", `Authorization: token 3be14d5fad7e7889485c40987d5185c2d21a187b`, `-XGET`,
				`https://api.github.com/search/repositories?s=stars&o=desc&q=language%3A` + lang + `+stars%3A<` + strconv.Itoa(maxStars) + `&type=Repositories&per_page=100&page=` + strconv.Itoa(i+1)}
			b, err := exec.Command("curl", args...).Output()
			if err != nil {
				log.Fatal(err)
			}

			var r ReposResult
			if err := json.Unmarshal(b, &r); err != nil {
				log.Fatal(err)
			}

			if i == 9 {
				if len(r.Items) == 0 {
					log.Printf("aborting lang %s on iter %d, i %d; last response: %s", lang, iter, i, string(b))
					return
				}
				nextMaxStars = r.Items[len(r.Items)-1].Stars
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

	wg.Wait()
	return
}

func main() {
	run()
}
