package main

import (
	"go.uber.org/atomic"
	"sync"
)

type Treasure string
type Coin int

type Treasurer struct {
	c               *Client
	coins           chan Coin
	treasuresToCash chan Treasure
	m               sync.RWMutex
	closeConnection *atomic.Bool

	addWorker    chan struct{}
	deleteWorker chan struct{}
}

func (t *Treasurer) cash(treasure Treasure) error {
	var coins []Coin
	_, err := t.c.doRequest("cash", &treasure, &coins, t.closeConnection.Load())

	if err != nil {
		t.treasuresToCash <- treasure
		return err
	}

	for _, coin := range coins {
		t.coins <- coin
	}

	return nil
}

func (t *Treasurer) GetCoins(number int) []Coin {
	// TODO: mojet zalochitsya
	coins := make([]Coin, 0, number)
	for i := 0; i < number; i++ {
		coins = append(coins, <-t.coins)
	}
	return coins
}

func (t *Treasurer) ReturnCoins(coins []Coin) {
	for _, coin := range coins {
		t.coins <- coin
	}
}

func (t *Treasurer) GetCoinsCount() int {
	return len(t.coins)
}

func (t *Treasurer) run() {
	for {
		<-t.addWorker
		go func() {
			for treasure := range t.treasuresToCash{
				t.cash(treasure)
			}
		}()
	}
}

func (t *Treasurer) Cash(treasure Treasure) {
	t.treasuresToCash <- treasure
}

func (t *Treasurer) Close() {
	t.closeConnection.Store(true)
}

func (t *Treasurer) AddWorkers(n int) {
	for i := 0; i < n; i++ {
		t.addWorker <- struct{}{}
	}
}

func (t *Treasurer) DeleteWorkers(n int) {
	for i := 0; i < n; i++ {
		t.deleteWorker <- struct{}{}
	}
}
func NewTreasurer(client *Client) *Treasurer {
	//client.SetRPSLimit("cash", 105)
	treasurer := Treasurer{
		c:               client,
		coins:           make(chan Coin, 1000000),
		treasuresToCash: make(chan Treasure, 1000000),
		closeConnection: atomic.NewBool(false),
		addWorker:    make(chan struct{}, 100),
		deleteWorker: make(chan struct{}, 100),
	}
	go treasurer.run()
	return &treasurer
}
