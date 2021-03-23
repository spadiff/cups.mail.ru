package main

import (
	"fmt"
	"sync"
)

var kek bool

type Licenser struct {
	c             *Client
	t             *Treasurer
	licenses      map[int]int
	licensesQueue chan int
	stat          map[int][]int
	now           int
	m             sync.RWMutex
}

func (l *Licenser) create(coins []Coin) (int, int, error) {
	response := struct {
		ID         int `json:"id"`
		DigAllowed int `json:"digAllowed"`
		DigUsed    int `json:"digUsed"` // TODO: is it really useless?
	}{}

	code, err := l.c.doRequest("licenses", &coins, &response)

	if code == 402 {
		if !kek {
			fmt.Println(err)
			kek = true
		}
		return 0, 0, nil
	}

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
		l.t.m.Lock()
		coins := []Coin{}
		coinsCount := l.t.GetCoinsCount()
		willUse := 12

		if willUse > coinsCount {
			willUse = 0
		}
		coins = l.t.GetCoins(willUse)
		l.t.m.Unlock()

		id, count, err := l.create(coins)
		if err != nil {
			l.t.m.Lock()
			l.t.ReturnCoins(coins)
			l.t.m.Unlock()
			continue
		}

		l.m.Lock()
		l.licenses[id] = count
		tries, ok := l.stat[willUse]
		if !ok {
			l.stat[willUse] = make([]int, 0)
		}
		if len(tries) >= 9 || count == 0 {
			l.now++
		}
		l.stat[willUse] = append(tries, count)
		l.m.Unlock()

		for i := 0; i < count; i++ {
			l.licensesQueue <- 1
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
		stat:          make(map[int][]int),
	}
	for i := 0; i < 10; i++ {
		go licenser.run()
	}
	return &licenser
}
