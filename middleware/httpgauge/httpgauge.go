//go:build !solution

package httpgauge

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type Gauge struct {
	stats map[string]int
	mu    sync.RWMutex
}

func New() *Gauge {
	return &Gauge{
		stats: make(map[string]int),
		mu:    sync.RWMutex{},
	}
}

func (g *Gauge) Snapshot() map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[string]int)
	maps.Copy(out, g.stats)
	return out
}

func (g *Gauge) inc(rp string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stats[rp]++
}

// ServeHTTP returns accumulated statistics in text format ordered by pattern.
//
// For example:
//
//	/a 10
//	/b 5
//	/c/{id} 7
func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sb := strings.Builder{}
	stats := g.Snapshot()
	for _, k := range slices.Sorted(maps.Keys(stats)) {
		v := stats[k]
		sb.WriteString(fmt.Sprintf("%s %d\n", k, v))
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(sb.String())); err != nil {
		panic(err)
	}
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rp := chi.RouteContext(r.Context()).RoutePattern()
			g.inc(rp)
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}
