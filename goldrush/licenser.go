package main

import (
	"sync"
)

type Licenser struct {
	c             *Client
	t             *Treasurer
	licenses      map[int]int
	licensesQueue chan int
	stat          map[int]int
	now           int
	nowTime       int
	m             sync.RWMutex
}

func (l *Licenser) create(coins []Coin) (int, int, error) {
	response := struct {
		ID         int `json:"id"`
		DigAllowed int `json:"digAllowed"`
		DigUsed    int `json:"digUsed"` // TODO: is it really useless?
	}{}

	_, err := l.c.doRequest("licenses", &coins, &response)
	if err != nil {
		return 0, 0, err
	}

	return response.ID, response.DigAllowed, nil
}

func (l *Licenser) GetLicense() int {
	<-l.licensesQueue

	l.m.Lock()
	defer l.m.Unlock()

	for k := range l.licenses {
		count := l.licenses[k] - 1
		if count == 0 {
			delete(l.licenses, k)
		} else {
			l.licenses[k] = count
		}
		return k
	}

	return 0
}

func (l *Licenser) ReturnLicense(k int) {
	l.m.Lock()
	count, ok := l.licenses[k]
	if !ok {
		count = 0
	}
	l.licenses[k] = count + 1
	l.m.Unlock()
	l.licensesQueue <- 1
}

func (l *Licenser) run() {
	for {
		coins := []Coin{}
		//coinsCount := l.t.GetCoinsCount()
		//if l.now >= coinsCount {
		//	coins = l.t.GetCoins(l.now)
		//}

		id, count, err := l.create(coins)
		if err == nil {
			l.m.Lock()
			l.licenses[id] = count
			l.stat[l.now] += count
			l.nowTime++
			if l.nowTime >= 9 {
				l.stat[l.now] /= l.nowTime
				l.nowTime = 0
				l.now++
			}
			l.m.Unlock()
			for i := 0; i < count; i++ {
				l.licensesQueue <- 1
			}
		}
	}
}

func NewLicenser(client *Client, treasurer *Treasurer) *Licenser {
	client.SetRPSLimit("licenses", 99)
	licenser := Licenser{
		c:             client,
		t:             treasurer,
		licenses:      make(map[int]int),
		licensesQueue: make(chan int, 100000),
		stat:          make(map[int]int),
		now:           0,
		nowTime: 0,
	}
	for i := 0; i < 10; i++ {
		go licenser.run()
	}
	return &licenser
}
