package main

import (
	"fmt"
	"time"
)

func main() {
	client := NewClient()
	treasurer := NewTreasurer(client)
	licenser := NewLicenser(client, treasurer)
	digger := NewDigger(client, licenser, treasurer)
	explorer := NewExplorer(client, digger)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for _ = range ticker.C {
			licenser.m.RLock()
			client.m.RLock()
			fmt.Printf(
				"l: %v, d: %v, ea: %d/%d\nq: %v\ns: %v\n",
				len(licenser.licenses),
				digger.pointsInQueue,
				explorer.emptyAreasCount,
				explorer.areasCount,
				client.queue,
				client.statuses,
			)
			licenser.m.RUnlock()
			client.m.RUnlock()
		}
	}()

	for i := 0; i < 10; i++ {
		go explorer.Run(350 * i, 350 * (i + 1), 5)
	}

	time.Sleep(10 * time.Minute)
}
