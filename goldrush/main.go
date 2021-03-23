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
		ticker := time.NewTicker(60 * time.Second)
		for _ = range ticker.C {
			licenser.m.RLock()
			client.m.RLock()
			fmt.Printf(
				"l: %v, d: %v\nq: %v\ns: %v\nl: %v\n",
				len(licenser.licenses),
				digger.pointsInQueue,
				client.queue,
				client.statuses,
				licenser.stat,
			)
			licenser.m.RUnlock()
			client.m.RUnlock()
		}
	}()

	explorer.Run()
}
