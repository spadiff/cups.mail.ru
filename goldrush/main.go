package main

import (
	"fmt"
	"go.uber.org/atomic"
	"sort"
	"time"
)

type Measure struct {
	stats map[string]*atomic.Int32
}

func NewMeasure(values []string) *Measure {
	measure := Measure{
		stats: make(map[string]*atomic.Int32),
	}
	for _, value := range values {
		measure.stats[value] = atomic.NewInt32(0)
	}
	return &measure
}

func (m *Measure) Add(name string, n int32) {
	m.stats[name].Add(n)
}

func (m *Measure) String() string {
	keys := make([]string, 0, len(m.stats))
	for key := range m.stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := ""
	for _, key := range keys {
		value := m.stats[key]
		if value.Load() != 0 {
			result += key + ": " + value.String() + ", "
		}
	}

	return result
}

func main() {
	client := NewClient()
	treasurer := NewTreasurer(client)
	licenser := NewLicenser(client, treasurer)
	digger := NewDigger(client, licenser, treasurer)
	explorer := NewExplorer(client, digger)

	go func() {
		time.Sleep(9*time.Minute + 30*time.Second)
		ticker := time.NewTicker(10 * time.Second)
		for _ = range ticker.C {
			fmt.Printf(
				"d: %v\n"+
				"e: %v\n" +
				"c: %v\n",
				digger.measure.String(),
				explorer.measure.String(),
				client.measure.String(),
			)
		}
	}()

	for i := 0; i < 10; i++ {
		go explorer.Run(350*i, 350*(i+1), 1024)
	}

	time.Sleep(10 * time.Minute)
}
