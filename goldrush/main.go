package main

import (
	"fmt"
	"time"
)

func main() {
	client := NewClient()
	licenser := NewLicenser(client)
	treasurer := NewTreasurer(client)
	digger := NewDigger(client, licenser, treasurer)
	explorer := NewExplorer(client, digger)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for _ = range ticker.C {
			fmt.Printf(
				"l: %d, t: %d, d: %d",
				len(licenser.licensesQueue),
				len(treasurer.treasuresToCash),
				len(digger.pointsToFind),
			)
		}
	}()

	explorer.Run()
}
