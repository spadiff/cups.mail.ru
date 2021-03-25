package main

import (
	"go.uber.org/atomic"
	"strconv"
	"sync"
	"time"
)

type Licenser struct {
	c                    *Client
	t                    *Treasurer
	licensesQueue        chan int
	licensesBeforePlatit *atomic.Int32
	m                    sync.RWMutex
	stop                 *atomic.Bool
	measure              *Measure
	cost                 int
	requests             *atomic.Int32
}

func (l *Licenser) create(coins []Coin) (int, int, error) {
	response := struct {
		ID         int `json:"id"`
		DigAllowed int `json:"digAllowed"`
		DigUsed    int `json:"digUsed"` // TODO: is it really useless?
	}{}

	_, err := l.c.doRequest("licenses", &coins, &response, false)

	if err != nil {
		return 0, 0, err
	}

	return response.ID, response.DigAllowed, nil
}

func (l *Licenser) GetLicense(d *Digger) int {
	before := time.Now()
	license := <-l.licensesQueue
	d.measure.Add("wait_license_count", 1)
	d.measure.Add("wait_license_time", time.Now().Sub(before).Microseconds())
	return license
}

func (l *Licenser) ReturnLicense(k int) {
	l.licensesQueue <- k
}

func (l *Licenser) Stop() {
	l.stop.Store(true)
}

func (l *Licenser) run() {
	for {
		if l.stop.Load() {
			break
		}

		//willUse := l.cost * int(l.requests.Load() / 30)
		willUse := l.cost

		if l.licensesBeforePlatit.Load() > 0 {
			willUse = 0
		}

		coins := l.t.GetCoins(willUse)

		id, count, err := l.create(coins)
		if err != nil {
			l.t.ReturnCoins(coins)
			continue
		}

		l.licensesBeforePlatit.Sub(1)
		l.measure.Add(strconv.Itoa(willUse) + "_count", 1)
		l.measure.Add(strconv.Itoa(willUse) + "_sum", int64(count))

		l.requests.Add(1)

		for i := 0; i < count; i++ {
			l.licensesQueue <- id
		}
	}
}

func NewLicenser(client *Client, treasurer *Treasurer, cost int) *Licenser {
	//client.SetRPSLimit("licenses", 25)

	measures := make([]string, 0)
	for i := 0; i <= 1000; i++ {
		measures = append(measures, strconv.Itoa(i) + "_count")
		measures = append(measures, strconv.Itoa(i) + "_sum")
	}

	licenser := Licenser{
		c:                    client,
		t:                    treasurer,
		licensesQueue:        make(chan int, 100000),
		licensesBeforePlatit: atomic.NewInt32(20),
		stop:                 atomic.NewBool(false),
		measure:              NewMeasure(measures),
		cost:                 cost,
		requests:             atomic.NewInt32(0),
	}

	return &licenser
}
