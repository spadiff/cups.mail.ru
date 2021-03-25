package main

import (
	"fmt"
	"go.uber.org/atomic"
	"runtime"
	"sort"
	"strconv"
	"sync"
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

func main() {
	client := NewClient()
	treasurer := NewTreasurer(client)
	licenser := NewLicenser(client, treasurer)
	digger := NewDigger(client, licenser, treasurer)
	explorer := NewExplorer(client, digger)

	fmt.Println("4 explore po 25, 2 lic, 5 dig, 1 + 1 treasurer (close), 3m explore, from 2m others, 12 monet")

	go func() {
		time.Sleep(9*time.Minute + 30*time.Second)
		ticker := time.NewTicker(30 * time.Second)
		for _ = range ticker.C {
			fmt.Printf(
				"d: %v\n"+
					"e: %v\n"+
					"c: %v\n"+
					"cash: %v\n",
				digger.measure.String(),
				explorer.measure.String(),
				client.measure.String(),
				len(treasurer.treasuresToCash),
			)
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for _ = range ticker.C {
			//buf := make([]byte, 1000000)
			//runtime.Stack(buf, true)
			fmt.Printf("g: %v\n", strconv.Itoa(runtime.NumGoroutine()))
		}
	}()

	now := time.Now()
	exploreBefore := now.Add(3*time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go explorer.Run(875*i, 875*(i+1), 25, exploreBefore, &wg)
	}
	//wg.Wait()

	time.Sleep(2*time.Minute)

	for i := 0; i < 2; i++ {
		go licenser.run()
	}

	digger.run()
	licenser.Stop()
	treasurer.Close()

	for i := 0; i < 3; i++ {
		go treasurer.run()
	}

	time.Sleep(10 * time.Minute)
}
