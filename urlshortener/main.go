//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/exp/rand"
	"io"
	"log"
	"net/http"
	"sync"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	rBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req shortenRequest
	if err := json.Unmarshal(rBytes, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("invalid request")); err != nil {
			log.Fatal(err)
		}
		return
	}

	key := generateKey(req.URL)
	key2urlMx.Lock()
	key2url[key] = req.URL
	key2urlMx.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := shortenResponse{
		URL: req.URL,
		Key: key,
	}
	respB, _ := json.Marshal(resp)
	if _, err := w.Write(respB); err != nil {
		log.Fatal(err)
	}
}

func goHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	key2urlMx.RLock()
	url, exists := key2url[key]
	key2urlMx.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("key not found")); err != nil {
			log.Fatal(err)
		}
		return
	}

	w.Header().Set("Location", url)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusFound)
	res := fmt.Sprintf("<a href=\"%s\">Found</a>.", url)
	if _, err := w.Write([]byte(res)); err != nil {
		log.Fatal(err)
	}
}

var key2url = make(map[string]string)
var key2urlMx = sync.RWMutex{}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "port to run server on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", shortenHandler)
	mux.HandleFunc("/go/{key}", goHandler)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

var alphabet = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

func generateKey(url string) string {
	rng := rand.New(rand.NewSource(seedFromString(url)))
	b := make([]rune, 6)
	for i := range b {
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	return string(b)
}

func seedFromString(s string) uint64 {
	var seed uint64
	for _, char := range s {
		seed = (seed << 5) + seed + uint64(char)
	}
	return seed
}
