//go:build !solution

package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	wordCount := make(map[string]int)

	for _, filePath := range os.Args[1:] {
		data, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		dataS := string(data)
		for _, word := range strings.Split(dataS, "\n") {
			wordCount[word]++
		}
	}

	for word, cnt := range wordCount {
		if cnt > 1 {
			fmt.Printf("%d\t%s\n", cnt, word)
		}
	}
}
