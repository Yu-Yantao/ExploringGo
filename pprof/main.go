package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"sort"
	"strings"
	"time"
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		log.Println("pprof server listening on :6060")
		log.Println("open http://localhost:6060/debug/pprof/")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
	}()

	for round := 1; ; round++ {
		start := time.Now()
		result := runWorkload(6000, rng)
		fmt.Printf("round=%d result=%d cost=%s\n", round, result, time.Since(start))
		time.Sleep(800 * time.Millisecond)
	}
}

func runWorkload(size int, rng *rand.Rand) int {
	texts := make([]string, 0, size)
	for i := 0; i < size; i++ {
		texts = append(texts, buildText(i, rng))
	}

	scores := make([]int, 0, len(texts))
	for _, text := range texts {
		scores = append(scores, scoreText(text))
	}

	sort.Ints(scores)

	total := 0
	for _, score := range scores {
		total += score
	}
	return total
}

func buildText(i int, rng *rand.Rand) string {
	parts := make([]string, 0, 80)
	for j := 0; j < 80; j++ {
		parts = append(parts, fmt.Sprintf("item-%d-%d-%d", i, j, rng.Intn(1000)))
	}
	return strings.Join(parts, ",")
}

func scoreText(text string) int {
	fields := strings.Split(text, ",")
	total := 0
	for _, field := range fields {
		for _, ch := range field {
			total += int(ch)
		}
	}
	return total
}
