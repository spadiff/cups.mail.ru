package main

import (
	"sync"
	"time"
)

type License struct {
	id   int
	digs int
}

type Licenser struct {
	c             *Client
	licenses      []License
	licensesQueue chan int
	m             sync.Mutex
}

func (l *Licenser) create(coins []Coin) (License, error) {
	response := struct {
		ID         int `json:"id"`
		DigAllowed int `json:"digAllowed"`
		DigUsed    int `json:"digUsed"` // TODO: is it really useless?
	}{}

	_, err := l.c.doRequest("licenses", &coins, &response)
	if err != nil {
		return License{}, err
	}

	return License{id: response.ID, digs: response.DigAllowed}, nil
}

func (l *Licenser) GetLicense() int {
	return <-l.licensesQueue
}

func (l *Licenser) ReturnLicense(id int) {
	l.licensesQueue <- id
}

func (l *Licenser) addToQueue(license License) {
	l.m.Lock()
	l.licenses = append(l.licenses, license)
	l.m.Unlock()

	for i := 0; i < license.digs; i++ {
		l.licensesQueue <- license.id
	}
}

func (l *Licenser) run() {
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		if len(l.licensesQueue) > 27 {
			continue
		}
		license, err := l.create([]Coin{})
		if err == nil {
			go l.addToQueue(license)
		}
	}
}

func NewLicenser(client *Client) *Licenser {
	licenser := Licenser{
		c:             client,
		licenses:      make([]License, 0),
		licensesQueue: make(chan int, 100000),
	}
	go licenser.run()
	return &licenser
}
