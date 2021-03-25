package main

import (
	"fmt"
	"go.uber.org/atomic"
	"sort"
	"time"
)

type Measure struct {
	stats map[string]*atomic.Int64
}

func NewMeasure(values []string) *Measure {
	measure := Measure{
		stats: make(map[string]*atomic.Int64),
	}
	for _, value := range values {
		measure.stats[value] = atomic.NewInt64(0)
	}
	return &measure
}

func (m *Measure) Add(name string, n int64) {
	m.stats[name].Add(n)
	return
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

func stat(
	digger *Digger,
	explorer *Explorer,
	client *Client,
	treasurer *Treasurer,
	licenser *Licenser,
) {
	time.Sleep(9*time.Minute + 30*time.Second)
	ticker := time.NewTicker(30 * time.Second)
	for _ = range ticker.C {
		fmt.Printf(
			"d: %v\n"+
				"e: %v\n"+
				"c: %v\n"+
				"l: %v\n"+
				"cash: %v\n",
			digger.measure.String(),
			explorer.measure.String(),
			client.measure.String(),
			licenser.measure.String(),
			len(treasurer.treasuresToCash),
		)
	}
}

const (
	shouldFind            = 27000
	exploreWorkers        = 8
	exploreWidth          = 25
	licenseWorkers        = 1
	licenseCost           = 12
	diggerWorkers         = 10
	treasureWorkers       = 4
	treasureFinishWorkers = 4
)

func main() {
	client := NewClient()
	treasurer := NewTreasurer(client)
	licenser := NewLicenser(client, treasurer, licenseCost)
	digger := NewDigger(client, licenser, treasurer)
	explorer := NewExplorer(client, digger, exploreWorkers, exploreWidth, shouldFind)

	fmt.Println("exp: lic in dig")
	fmt.Printf(
		"%d explore po %d, %d license %d monet, %d dig, %d -> %d treasurer, goal %d\n",
		exploreWorkers,
		exploreWidth,
		licenseWorkers,
		licenseCost,
		diggerWorkers,
		treasureWorkers,
		treasureFinishWorkers,
		shouldFind,
	)

	go stat(digger, explorer, client, treasurer, licenser)

	time.Sleep(15 * time.Minute)
}
