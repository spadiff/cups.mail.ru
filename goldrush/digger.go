package main

import (
	"fmt"
	"go.uber.org/atomic"
	"strconv"
	"strings"
	"time"
)

type Digger struct {
	c            *Client
	l            *Licenser
	t            *Treasurer
	pointsToFind chan Point
	measure      *Measure

	workers      *atomic.Int32
	addWorker    chan struct{}
	deleteWorker chan struct{}
}

func (d *Digger) dig(point Point, depth int, license int) ([]Treasure, error) {
	request := struct {
		X         int `json:"posX"`
		Y         int `json:"posY"`
		Depth     int `json:"depth"`
		LicenseID int `json:"licenseID"`
	}{
		X: point.x, Y: point.y, Depth: depth, LicenseID: license,
	}

	response := make([]Treasure, 0)

	code, err := d.c.doRequest("dig", &request, &response, false)
	if err != nil && code != 404 {
		if strings.Contains(err.Error(), "no such license") {
			fmt.Println("no license, status code "+strconv.Itoa(code), license)
			return response, nil
		}
		return nil, err
	}

	return response, nil
}

func (d *Digger) run() {
	for i := 0; i < 10; i++ {

		go func() {
			<-d.addWorker
			before := time.Now()
			license, count := 0, 0
			for {
				done := false

				select {
				case point, ok := <-d.pointsToFind:
					if !ok {
						done = true
						d.l.Stop()
						fmt.Println("digger:", time.Now().Sub(before).Seconds())
						d.t.SetWorkers(treasureFinishWorkers)
						break
					}

					d.measure.Add("points_queue", 1)

					for depth := 1; depth <= MAX_DEPTH; depth++ {

						if count == 0 {
							var err error
							willUse := d.l.cost

							if d.l.licensesBeforePlatit.Load() > 0 {
								willUse = 0
							}
							coins := d.l.t.GetCoins(willUse)
							license, count, err = d.l.create(coins)
							for err != nil {
								license, count, err = d.l.create(coins)
							}
							d.l.licensesBeforePlatit.Sub(1)
						}

						//license := d.l.GetLicense(d)
						treasures, err := d.dig(point, depth, license)
						if err != nil {
							//d.l.ReturnLicense(license)
							return
						}
						count -= 1

						d.measure.Add("depth_"+strconv.Itoa(depth)+"_sum", int64(len(treasures)))
						for _, treasure := range treasures {
							d.t.Cash(treasure)
						}
						point.amount -= len(treasures)
						if point.amount <= 0 {
							break
						}
					}

					d.measure.Add("points_queue", -1)
				case <-d.deleteWorker:
					<-d.addWorker
				}

				if done {
					break
				}
			}
		}()
	}
}

func (d *Digger) SetWorkers(n int) {
	workers := int(d.workers.Load())
	if workers > n {
		for i := 0; i < workers-n; i++ {
			d.deleteWorker <- struct{}{}
		}
	} else if workers < n {
		for i := 0; i < n-workers; i++ {
			d.addWorker <- struct{}{}
		}
	}
	d.workers.Store(int32(n))
}

func (d *Digger) Find(point Point) {
	d.pointsToFind <- point
}

func (d *Digger) Done() {
	//close(d.pointsToFind)
	//for i := 0; i < licenseWorkers; i++ {
	//	go d.l.run()
	//}

	d.SetWorkers(diggerWorkers)
	d.t.SetWorkers(treasureWorkers)

	//kek := false
	//
	//for {
	//	if len(d.t.treasuresToCash) > 2000 {
	//		d.t.SetWorkers(6)
	//		d.SetWorkers(0)
	//		kek = true
	//	} else if len(d.t.treasuresToCash) < 50 && kek {
	//		d.SetWorkers(10)
	//		d.t.SetWorkers(0)
	//		kek = false
	//	}
	//}
}

func NewDigger(client *Client, licenser *Licenser, treasurer *Treasurer) *Digger {
	//client.SetRPSLimit("dig", 499)
	measure := []string{"points_queue", "wait_license_count", "wait_license_time"}
	for i := 1; i <= 10; i++ {
		measure = append(measure, "depth_"+strconv.Itoa(i)+"_sum")
	}
	digger := Digger{
		c:            client,
		l:            licenser,
		t:            treasurer,
		pointsToFind: make(chan Point, 1000000),
		measure:      NewMeasure(measure),
		workers:      atomic.NewInt32(0),
		addWorker:    make(chan struct{}, 100),
		deleteWorker: make(chan struct{}, 100),
	}
	go digger.run()
	return &digger
}
