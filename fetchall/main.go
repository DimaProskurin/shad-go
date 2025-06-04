//go:build !solution

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func fetchURL(url string, results chan<- string) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		results <- err.Error()
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		results <- err.Error()
		return
	}

	dur := time.Since(start)
	results <- fmt.Sprintf("url %s result len %d; time %v", url, len(string(body)), dur)
}

func main() {
	results := make(chan string)
	urlsCnt := 0

	startTime := time.Now()
	for _, url := range os.Args[1:] {
		go fetchURL(url, results)
		urlsCnt++
	}

	for i := 0; i < urlsCnt; i++ {
		res := <-results
		fmt.Println(res)
	}

	fmt.Println(time.Since(startTime))
}
