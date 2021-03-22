package main

import (
	"sync"
	"time"
)

type Licenser struct {
	c             *Client
	licenses      map[int]int
	licensesQueue chan int
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
	for {
		l.m.RLock()
		if len(l.licenses) != 0 {
			l.m.RUnlock()
			break
		}
		l.m.RUnlock()
		time.Sleep(time.Millisecond)
	}

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
}

func (l *Licenser) run() {
	for {
		if len(l.licenses) >= 10 {
			continue
		}
		id, count, err := l.create([]Coin{})
		if err == nil {
			l.m.Lock()
			l.licenses[id] = count
			l.m.Unlock()
		}
	}
}

func NewLicenser(client *Client) *Licenser {
	client.SetRPSLimit("licenses", 99)
	licenser := Licenser{
		c:             client,
		licenses:      make(map[int]int),
		licensesQueue: make(chan int, 100000),
	}
	go licenser.run()
	return &licenser
}
